package templateService

import (
	"context"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	tmplRank "github.com/ZiRunHua/LeapLedger/service/template/rank"
	"strconv"
	"time"
)

var rank tmplRank.Rank[rankMember]

func initRank() {
	list, err := TemplateApp.GetList()
	if err != nil {
		panic(err)
	}
	members := make([]rankMember, len(list), len(list))
	for i, account := range list {
		members[i] = newRankMember(account)
	}
	rank = tmplRank.NewRank[rankMember]("tmplAccount", members, time.Hour*24)
	err = rank.Init(context.TODO())
	if err != nil {
		panic(err)
	}
}

type rankMember struct {
	tmplRank.Member
	key string
	id  uint
}

func newRankMember(account accountModel.Account) rankMember {
	return rankMember{id: account.ID, key: strconv.Itoa(int(account.ID))}
}

func (rm rankMember) String() string {
	return rm.key
}
