package response

import (
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/dataTool"
	"time"
)

// swagger:response// swagger:response
type CommonCaptcha struct {
	CaptchaId     string
	PicBase64     string
	CaptchaLength int
	OpenCaptcha   bool
}

// swagger:response
type Id struct {
	Id uint
}

// swagger:response
type CreateResponse struct {
	Id        uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

// swagger:response
type Token struct {
	Token               string
	TokenExpirationTime time.Time
}

// swagger:response
type TwoLevelTree struct {
	Tree []Father
}

// swagger:response
type Father struct {
	NameId
	Children []NameId
}

// swagger:response
type NameId struct {
	Id   uint
	Name string
}

// swagger:response
type NameValue struct {
	Name  string
	Value int
}

// swagger:response
type PageData struct {
	page  int
	limit int
	count int
}

// swagger:response
type ExpirationTime struct {
	ExpirationTime int
}

// swagger:model List
type List[T any] struct {
	// discriminator: true
	List []T
}

func getUsernameMap(ids []uint) (map[uint]string, error) {
	var nameList dataTool.Slice[uint, struct {
		ID       uint
		Username string
	}]
	err := db.Db.Model(&userModel.User{}).Where("id IN (?)", ids).Find(&nameList).Error
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
	err := db.Db.Model(&accountModel.Account{}).Where("id IN (?)", ids).Find(&nameList).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string)
	for _, s := range nameList {
		result[s.ID] = s.Name
	}
	return result, nil
}
