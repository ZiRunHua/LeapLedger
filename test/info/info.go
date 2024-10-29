package info

import "encoding/json"

var (
	Data Info
)

type Info struct {
	UserId    uint
	Email     string
	AccountId uint
	Token     string
}

func (i *Info) ToString() string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return string(b)
}
