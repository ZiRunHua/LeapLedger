package ws

import (
	"KeepAccount/api/response"
	"KeepAccount/api/v1/ws/msg"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	transactionModel "KeepAccount/model/transaction"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"io"
	"sync"
	"sync/atomic"
)

type BillImportWebsocket struct {
	account accountModel.Account
	conn    *websocket.Conn
	msg.Reader

	lock           sync.Mutex
	WaitRetryTrans map[string]transactionModel.Info
	RetryingTrans  map[string]transactionModel.Info

	total TotalData
}

func NewBillImportWebsocket(conn *websocket.Conn, account accountModel.Account) *BillImportWebsocket {
	return &BillImportWebsocket{
		account: account,
		conn:    conn,
		Reader:  msg.NewReader(),
	}
}

func (b *BillImportWebsocket) SendTransactionCreateSuccess(transaction transactionModel.Transaction) error {
	var transDetail response.TransactionDetail
	err := transDetail.SetData(transaction, &b.account)
	if err != nil {
		return err
	}
	b.total.add(transaction.IncomeExpense, transaction.Amount)
	return msg.Send(b.conn, "createSuccess", transDetail)
}

func (b *BillImportWebsocket) SendTransactionCreateFail(transInfo transactionModel.Info, failErr error) error {
	var transDetail response.TransactionDetail
	err := transDetail.SetDataIgnoreErr(transactionModel.Transaction{Info: transInfo}, &b.account)
	if err != nil {
		return err
	}
	id := uuid.NewString()
	msgHandler := func() {
		b.lock.Lock()
		defer b.lock.Unlock()
		if b.WaitRetryTrans == nil {
			b.WaitRetryTrans = make(map[string]transactionModel.Info)
		}
		b.WaitRetryTrans[id] = transInfo
	}
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
	msgHandler()
	return nil
}

func (b *BillImportWebsocket) RegisterMsgHandlerCreateRetry(handler func(transactionModel.Info) error) {
	type MsgTransactionCreateRetry struct {
		Id        string
		TransInfo transactionModel.Info
	}
	msg.RegisterHandle[MsgTransactionCreateRetry](b.Reader, "createRetry",
		func(data MsgTransactionCreateRetry) (err error) {
			mapHandler := func() error {
				b.lock.Lock()
				defer b.lock.Unlock()
				if _, exist := b.WaitRetryTrans[data.Id]; !exist {
					return msg.SendError(b.conn, global.ErrOperationTooFrequent)
				}
				delete(b.WaitRetryTrans, data.Id)
				if b.RetryingTrans == nil {
					b.RetryingTrans = make(map[string]transactionModel.Info)
				}
				b.RetryingTrans[data.Id] = data.TransInfo
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
				b.lock.Lock()
				defer b.lock.Unlock()
				delete(b.RetryingTrans, data.Id)
				if err == nil {
					err = b.tryFinish()
				}
			}()
			return nil
		})
}

func (b *BillImportWebsocket) RegisterMsgHandlerIgnoreTrans() {
	msg.RegisterHandle[string](b.Reader, "ignoreTrans",
		func(id string) (err error) {
			b.lock.Lock()
			defer b.lock.Unlock()
			if _, exist := b.WaitRetryTrans[id]; !exist {
				return msg.SendError(b.conn, global.ErrOperationTooFrequent)
			}
			delete(b.WaitRetryTrans, id)
			b.total.ignore()
			return b.tryFinish()
		})
}

func (b *BillImportWebsocket) TryFinish() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.tryFinish()
}

func (b *BillImportWebsocket) tryFinish() error {
	if len(b.WaitRetryTrans) != 0 || len(b.RetryingTrans) != 0 {
		return nil
	}
	return b.SendFinish()
}

func (b *BillImportWebsocket) SendFinish() error {
	type Total struct {
		ExpenseAmount, IncomeAmount            int64
		ExpenseCount, IncomeCount, IgnoreCount int32
	}
	return msg.Send[Total](b.conn, "finish", Total{
		ExpenseAmount: b.total.ExpenseAmount.Load(),
		IncomeAmount:  b.total.IncomeAmount.Load(),
		ExpenseCount:  b.total.ExpenseCount.Load(),
		IncomeCount:   b.total.IncomeCount.Load(),
		IgnoreCount:   b.total.IgnoreCount.Load(),
	})
}

func (b *BillImportWebsocket) Read() error {
	return msg.ForReadAndHandleJsonMsg(b.Reader, b.conn)
}

func (b *BillImportWebsocket) ReadFile() (name []byte, file io.Reader, err error) {
	name, err = msg.ReadBytes(b.Reader, b.conn)
	if err != nil {
		return
	}
	file, err = msg.ReadFile(b.Reader, b.conn)
	return
}

func (b *BillImportWebsocket) SendError() error {
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
