package ws

import (
	"errors"
	"io"
	"sync/atomic"

	"github.com/ZiRunHua/LeapLedger/api/response"
	"github.com/ZiRunHua/LeapLedger/api/v1/ws/msg"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	"github.com/ZiRunHua/LeapLedger/util/dataTool"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type (
	BillImportWebsocket interface {
		Read() error
		ReadFile() (name []byte, file io.Reader, err error)

		SendTransactionCreateSuccess(transaction transactionModel.Transaction) error
		SendTransactionCreateFail(transInfo transactionModel.Info, failErr error) error

		RegisterMsgHandlerCreateRetry(handler func(transactionModel.Info) error)
		RegisterMsgHandlerIgnoreTrans()
		RegisterMsgHandlerConfigUpdate(afterConfigChange func(config userModel.BillImportConfig) error)

		GetConfig() userModel.BillImportConfig
		TryFinish() error
		SendError() error
	}

	billImportWebsocket struct {
		account     accountModel.Account
		accountUser accountModel.User
		Config      BillImportConfig

		conn *websocket.Conn
		msg.Reader

		WaitRetryTrans *dataTool.RWMutexMap[string, transactionModel.Info]
		RetryingTrans  *dataTool.RWMutexMap[string, transactionModel.Info]

		total TotalData
	}
)

func NewBillImportWebsocket(conn *websocket.Conn, account accountModel.Account, accountUser accountModel.User) (
	BillImportWebsocket, error,
) {
	billConfig := &userModel.BillImportConfig{}
	billConfig.SetUserId(accountUser.UserId)
	err := userModel.NewDao().GetConfig(billConfig)
	if err != nil {
		return nil, err
	}
	return &billImportWebsocket{
		account:        account,
		Config:         BillImportConfig{config: billConfig},
		conn:           conn,
		Reader:         msg.NewReader(),
		WaitRetryTrans: dataTool.NewRWMutexMap[string, transactionModel.Info](),
		RetryingTrans:  dataTool.NewRWMutexMap[string, transactionModel.Info](),
	}, nil
}

func (b *billImportWebsocket) SendTransactionCreateSuccess(transaction transactionModel.Transaction) error {
	var transDetail response.TransactionDetail
	err := transDetail.SetData(transaction, &b.account)
	if err != nil {
		return err
	}
	b.total.add(transaction.IncomeExpense, transaction.Amount)
	return msg.Send(b.conn, "createSuccess", transDetail)
}

func (b *billImportWebsocket) SendTransactionCreateFail(transInfo transactionModel.Info, failErr error) error {
	var transDetail response.TransactionDetail
	err := transDetail.SetDataIgnoreErr(
		transactionModel.Transaction{
			Info:       transInfo,
			RecordType: transactionModel.RecordTypeOfImport,
		}, &b.account,
	)
	if err != nil {
		return err
	}
	id := uuid.NewString()
	type MsgTransactionCreateFail struct {
		Id    string
		Trans response.TransactionDetail
		Msg   string
	}
	err = msg.Send(
		b.conn,
		"createFail",
		MsgTransactionCreateFail{Id: id, Trans: transDetail, Msg: failErr.Error()},
	)
	if err != nil {
		return err
	}
	b.WaitRetryTrans.Store(id, transInfo)
	return nil
}

func (b *billImportWebsocket) RegisterMsgHandlerCreateRetry(handler func(transactionModel.Info) error) {
	type MsgTransactionCreateRetry struct {
		Id        string
		TransInfo transactionModel.Info
	}
	msg.RegisterHandle[MsgTransactionCreateRetry](
		b.Reader, "createRetry",
		func(data MsgTransactionCreateRetry) (err error) {
			mapHandler := func() error {
				if _, exist := b.WaitRetryTrans.Load(data.Id); !exist {
					return msg.SendError(b.conn, global.ErrOperationTooFrequent)
				}
				b.WaitRetryTrans.Delete(data.Id)
				b.RetryingTrans.Store(data.Id, data.TransInfo)
				return nil
			}
			err = mapHandler()
			if err != nil {
				return err
			}
			err = handler(data.TransInfo)
			if err != nil {
				return err
			}
			defer func() {
				b.RetryingTrans.Delete(data.Id)
				if err == nil {
					err = b.tryFinish()
				}
			}()
			return nil
		},
	)
}

func (b *billImportWebsocket) RegisterMsgHandlerIgnoreTrans() {
	msg.RegisterHandle[string](
		b.Reader, "ignoreTrans",
		func(id string) (err error) {
			if _, exist := b.WaitRetryTrans.Load(id); !exist {
				return msg.SendError(b.conn, global.ErrOperationTooFrequent)
			}
			b.WaitRetryTrans.Delete(id)
			b.total.ignore()
			return b.tryFinish()
		},
	)
}

func (b *billImportWebsocket) RegisterMsgHandlerConfigUpdate(afterConfigChange func(config userModel.BillImportConfig) error) {
	updateHandle := func(f func() error) error {
		if err := f(); err != nil {
			return err
		}
		return afterConfigChange(*b.Config.config)
	}
	msg.RegisterHandle[[]byte](
		b.Reader, "configUpdateIgnoreSameTrans",
		func(_ []byte) error { return updateHandle(b.Config.tempIgnoreUnmappedCategory) },
	)
	msg.RegisterHandle[[]byte](
		b.Reader, "configUpdateIgnoreAllSameTrans",
		func(_ []byte) error { return updateHandle(b.Config.ignoreSameTransaction) },
	)
	msg.RegisterHandle[[]byte](
		b.Reader, "configUpdateIgnoreUnmappedCategory",
		func(_ []byte) error { return updateHandle(b.Config.tempIgnoreUnmappedCategory) },
	)
	msg.RegisterHandle[[]byte](
		b.Reader, "configUpdateIgnoreAllUnmappedCategory",
		func(_ []byte) error { return updateHandle(b.Config.ignoreSameTransaction) },
	)
}

func (b *billImportWebsocket) GetConfig() userModel.BillImportConfig {
	return *b.Config.config
}
func (b *billImportWebsocket) TryFinish() error {
	return b.tryFinish()
}

func (b *billImportWebsocket) tryFinish() error {
	if b.WaitRetryTrans.Len() != 0 || b.RetryingTrans.Len() != 0 {
		return nil
	}
	return b.SendFinish()
}

func (b *billImportWebsocket) SendFinish() error {
	type Total struct {
		ExpenseAmount, IncomeAmount            int64
		ExpenseCount, IncomeCount, IgnoreCount int32
	}
	return msg.Send[Total](
		b.conn, "finish", Total{
			ExpenseAmount: b.total.ExpenseAmount.Load(),
			IncomeAmount:  b.total.IncomeAmount.Load(),
			ExpenseCount:  b.total.ExpenseCount.Load(),
			IncomeCount:   b.total.IncomeCount.Load(),
			IgnoreCount:   b.total.IgnoreCount.Load(),
		},
	)
}

func (b *billImportWebsocket) Read() error {
	return msg.ForReadAndHandleJsonMsg(b.Reader, b.conn)
}

func (b *billImportWebsocket) ReadFile() (name []byte, file io.Reader, err error) {
	name, err = msg.ReadBytes(b.Reader, b.conn)
	if err != nil {
		return
	}
	file, err = msg.ReadFile(b.Reader, b.conn)
	return
}

func (b *billImportWebsocket) SendError() error {
	return msg.SendError(b.conn, errors.New("test"))
}

type TotalData struct {
	ExpenseAmount, IncomeAmount            atomic.Int64
	ExpenseCount, IncomeCount, IgnoreCount atomic.Int32
}

func (t *TotalData) add(ie constant.IncomeExpense, amount int) {
	if ie == constant.Income {
		t.IncomeAmount.Add(int64(amount))
		t.IncomeCount.Add(1)
	} else {
		t.ExpenseAmount.Add(int64(amount))
		t.ExpenseCount.Add(1)
	}
}

func (t *TotalData) ignore() {
	t.IgnoreCount.Add(1)
}

// BillImportConfig Manage config updates from clients when importing
type BillImportConfig struct {
	config *userModel.BillImportConfig
}

func (bic *BillImportConfig) ignoreUnmappedCategory() error {
	bic.config.IgnoreUnmappedCategory = true
	return userModel.NewDao().UpdateConfigColumns(bic.config, "ignore_unmapped_category")
}

func (bic *BillImportConfig) tempIgnoreUnmappedCategory() error {
	bic.config.CheckSameTransMode = userModel.CheckSameTransModeOfIgnore
	return nil
}

func (bic *BillImportConfig) ignoreSameTransaction() error {
	bic.config.CheckSameTransMode = userModel.CheckSameTransModeOfIgnore
	return userModel.NewDao().UpdateConfigColumns(bic.config, "check_same_trans_mode")
}

func (bic *BillImportConfig) tempIgnoreSameTransaction() error {
	bic.config.CheckSameTransMode = userModel.CheckSameTransModeOfIgnore
	return nil
}
