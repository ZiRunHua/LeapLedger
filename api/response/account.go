package response

import accountModel "KeepAccount/model/account"

func AccountModelToResponse(account *accountModel.Account) *AccountOne {
	if account == nil {
		return &AccountOne{}
	}
	return &AccountOne{
		Id:        account.ID,
		Name:      account.Name,
		Icon:      account.Icon,
		UpdatedAt: account.UpdatedAt.Unix(),
		CreatedAt: account.CreatedAt.Unix(),
	}
}

type AccountOne struct {
	Id        uint
	Name      string
	Icon      string
	CreatedAt int64
	UpdatedAt int64
}

type AccountGetAll struct {
	List []AccountOne
}

type AccountTemplateOne struct {
	Id   uint
	Name string
	Icon string
}

type AccountTemplateList struct {
	List []AccountTemplateOne
}
