package response

import accountModel "KeepAccount/model/account"

func AccountModelToResponse(account *accountModel.Account) *AccountOne {
	return &AccountOne{
		Id:        account.ID,
		Name:      account.Name,
		UpdatedAt: account.UpdatedAt.Unix(),
		CreatedAt: account.CreatedAt.Unix(),
	}
}

type AccountOne struct {
	Id        uint
	Name      string `binding:"required"`
	CreatedAt int64
	UpdatedAt int64
}

type AccountGetOne struct {
	AccountOne
}

type AccountGetAll struct {
	List []AccountOne
}
