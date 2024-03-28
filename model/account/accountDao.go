package accountModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type AccountDao struct {
	db *gorm.DB
}

func NewDao(db ...*gorm.DB) *AccountDao {
	if len(db) > 0 {
		return &AccountDao{
			db: db[0],
		}
	}
	return &AccountDao{global.GvaDb}
}

func (a *AccountDao) SelectById(id uint) (account Account, err error) {
	err = a.db.Model(&account).First(&account, id).Error
	return
}

func (a *AccountDao) Update(account Account, data AccountUpdateData) error {
	updateData, err := data.getAccount()
	if err != nil {
		return err
	}
	return a.db.Model(&account).Where("id = ?", account.ID).Updates(updateData).Error
}

func (a *AccountDao) SelectMappingById(id uint) (Mapping, error) {
	var mapping Mapping
	return mapping, a.db.First(&mapping, id).Error
}

func (a *AccountDao) SelectMappingByMainAccountAndRelatedUser(mainAccountId, userId uint) (result Mapping, err error) {
	err = a.db.Model(&Mapping{}).Select("account_mapping.*").Joins("LEFT JOIN account ON account_mapping.related_id = account.id").Where(
		"account_mapping.main_id = ? AND account.user_id = ?", mainAccountId, userId,
	).First(&result).Error
	return
}

func (a *AccountDao) SelectAllMappingByAccount(account Account) ([]Mapping, error) {
	var result []Mapping
	err := a.db.Model(&Mapping{}).Where("main_id = ?", account.ID).Find(&result).Error
	return result, err
}

func (a *AccountDao) CreateMapping(mainAccount Account, mappingAccount Account) (Mapping, error) {
	mapping := Mapping{MainId: mainAccount.ID, RelatedId: mappingAccount.ID}
	return mapping, a.db.Create(&mapping).Error
}

func (a *AccountDao) DeleteMapping(mapping Mapping) error {
	return a.db.Delete(&mapping).Error
}

func (a *AccountDao) UpdateRelatedAccount(mapping Mapping, account Account) error {
	return a.db.Model(&mapping).Update("related_id", account.ID).Error
}

func (a *AccountDao) CreateUser(accountId uint, userId uint, permission UserPermission) (User, error) {
	data := User{
		AccountId:  accountId,
		UserId:     userId,
		Permission: permission,
	}
	return data, a.db.Create(&data).Error
}

func (a *AccountDao) UpdateUser(accountUser User, data UserUpdateData) (User, error) {
	err := a.db.Model(&accountUser).Update("permission", data.Permission).Error
	return accountUser, err
}

func (a *AccountDao) SelectUser(accountId uint, userId uint) (user User, err error) {
	err = a.db.Where("account_id = ? AND user_id = ?", accountId, userId).First(&user).Error
	return
}

func (a *AccountDao) ExistUser(accountId uint, userId uint) (exist bool, err error) {
	err = a.db.Raw(
		"SELECT EXISTS(SELECT 1 FROM account_user WHERE account_id = ? AND user_id = ?) AS exist", accountId, userId,
	).Scan(&exist).Error
	return
}

func (a *AccountDao) SelectUserListByAccountId(accountId uint) ([]User, error) {
	var result []User
	err := a.db.Model(&User{}).Where("account_id = ?", accountId).Order("id ASC").Find(&result).Error
	return result, err
}

func (a *AccountDao) SelectUserListByUserAndAccountType(userId uint, t Type) (result []User, err error) {
	query := a.db.Where("account_user.user_id = ? AND account.type = ?", userId, t)
	query = query.Select("account_user.*").Joins("LEFT JOIN account ON account.id = account_user.account_id")
	err = query.Order("account_user.id ASC").Find(&result).Error
	return
}

type UserCondition struct {
	t *Type
}

func NewUserCondition() *UserCondition {
	return &UserCondition{}
}
func (uc *UserCondition) SetType(t Type) *UserCondition {
	uc.t = &t
	return uc
}

func (uc *UserCondition) addConditionToQuery(db *gorm.DB) *gorm.DB {
	query := db
	if uc.t != nil {
		query = query.Where("account.type = ?", *uc.t)
	}
	return query
}

func (a *AccountDao) SelectByUserAndAccountType(userId uint, condition UserCondition) (result Account, err error) {
	query := a.db.Where("account_user.user_id = ?", userId)
	query = condition.addConditionToQuery(query)
	query = query.Select("account.*").Joins("LEFT JOIN account_user ON account.id = account_user.account_id")
	err = query.Order("account.id DESC").Find(&result).Error
	return
}

func (a *AccountDao) SelectUserInvitation(accountId uint, inviteeId uint) (invitation UserInvitation, err error) {
	err = a.db.Model(&invitation).Where("account_id = ? AND invitee = ?", accountId, inviteeId).First(&invitation).Error
	return
}

func (a *AccountDao) CreateUserInvitation(
	accountId uint, inviterUserId uint, inviteeUserId uint, permission UserPermission,
) (UserInvitation, error) {
	data := UserInvitation{
		AccountId:  accountId,
		Inviter:    inviterUserId,
		Invitee:    inviteeUserId,
		Status:     UserInvitationStatsOfWaiting,
		Permission: permission,
	}
	return data, a.db.Create(&data).Error
}

func (a *AccountDao) SelectUserInvitationByCondition(condition UserInvitationCondition) ([]UserInvitation, error) {
	query := a.db
	if condition.AccountId != nil {
		query = query.Where("account_id = ?", condition.AccountId)
	}
	if condition.InviterId != nil {
		query = query.Where("inviter = ?", condition.InviterId)
	}
	if condition.InviteeId != nil {
		query = query.Where("invitee = ?", condition.InviteeId)
	}
	if condition.Permission != nil {
		query = query.Where("permission = ?", condition.Permission)
	}
	var result []UserInvitation
	err := query.Limit(condition.Limit).Offset(condition.Offset).Order("id DESC").Find(&result).Error
	return result, err
}
