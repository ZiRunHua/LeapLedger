package response

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/dataTool"
)

type CommonCaptcha struct {
	CaptchaId     string
	PicBase64     string
	CaptchaLength int
	OpenCaptcha   bool
}
type Id struct {
	Id uint
}

type CreateResponse struct {
	Id        uint
	CreatedAt int64
	UpdatedAt int64
}

type Token struct {
	Token               string
	TokenExpirationTime int64
}
type TwoLevelTree struct {
	Tree []Father
}
type Father struct {
	NameId
	Children []NameId
}
type NameId struct {
	Id   uint
	Name string
}
type NameValue struct {
	Name  string
	Value int
}

type PageData struct {
	page  int
	limit int
	count int
}

type ExpirationTime struct {
	ExpirationTime int
}

type List[T any] struct {
	List []T
}

func getUsernameMap(ids []uint) (map[uint]string, error) {
	var nameList dataTool.Slice[uint, struct {
		ID       uint
		Username string
	}]
	err := global.GvaDb.Model(&userModel.User{}).Where("id IN (?)", ids).Find(&nameList).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string)
	for _, s := range nameList {
		result[s.ID] = s.Username
	}
	return result, nil
}

func getAccountNameMap(ids []uint) (map[uint]string, error) {
	var nameList dataTool.Slice[uint, struct {
		ID   uint
		Name string
	}]
	err := global.GvaDb.Model(&accountModel.Account{}).Where("id IN (?)", ids).Find(&nameList).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string)
	for _, s := range nameList {
		result[s.ID] = s.Name
	}
	return result, nil
}
