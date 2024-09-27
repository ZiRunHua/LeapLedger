package product

import (
	"KeepAccount/global/db"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	transactionModel "KeepAccount/model/transaction"
	test "KeepAccount/test/initialize"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Msg struct {
	Type string
	Data map[string]interface{}
}

func TestWebsocket(t *testing.T) {
	path := fmt.Sprintf(
		"ws://%s/account/%d/product/%s/bill/import",
		test.Host,
		test.Info.AccountId,
		productModel.AliPay)
	c, response, err := websocket.DefaultDialer.Dial(path,
		map[string][]string{"authorization": {test.Info.Token}})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response)
	defer c.Close()
	var amount = 0
	go func() {
		var msg Msg
		for {
			_, b, err := c.ReadMessage()
			if err != nil {
				t.Fatal(err)
				return
			}
			t.Log(msg)
			err = json.Unmarshal(b, &msg)
			if err != nil {
				t.Fatal(err)
				return
			}
			switch msg.Type {
			case "createSuccess":
				switch value := msg.Data["Amount"].(type) {
				case float64:
					amount += int(value)
				case int:
					amount += value
				case string:
					i, err := strconv.Atoi(value)
					if err != nil {
						t.Fatal(err)
					}
					amount += i
				}
			case "createFail":
				var category categoryModel.Category
				err = db.Db.First(&category).Error
				if err != nil {
					t.Fatal(err)
				}
				log.Print(b)
				var data MsgTransactionCreateFail
				err = json.Unmarshal([]byte(strings.TrimSpace(string(b))), &data)
				if err != nil {
					t.Fatal(err)
					return
				}
				data.Data.TransInfo.CategoryId = category.ID
				data.Data.TransInfo.IncomeExpense = category.IncomeExpense
				data.Type = "createRetry"
				err = c.WriteJSON(data)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}()
	fileName := "alipay_bill.csv"
	err = c.WriteMessage(websocket.TextMessage, []byte(fileName))
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("File read error:", err)
			return
		}

		err = c.WriteMessage(websocket.BinaryMessage, buffer[:n])
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}
	}
	err = c.WriteMessage(websocket.TextMessage, []byte("send finish"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("send finish")
	time.Sleep(time.Second * 10)
	if amount != 15902 {
		t.Fatal("err amount:", amount)
	}
}

type MsgTransactionCreateFail struct {
	Type string
	Data struct {
		Id        string
		TransInfo transactionModel.Info
		ErrStr    string
	}
}
