package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CategoryApi struct {
}

func (catApi *CategoryApi) CreateOne(ctx *gin.Context) {
	var requestData request.CategoryCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	father := categoryModel.Father{}
	err := father.SelectById(requestData.FatherId)
	if pass := checkFunc.AccountBelong(father.AccountId, ctx); false == pass {
		return
	}
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var category categoryModel.Category
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			category, err = categoryService.CreateOne(
				father,
				categoryService.NewCategoryData(categoryModel.Category{Name: requestData.Name, Icon: requestData.Icon}),
				tx,
			)
			return err
		},
	)

	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.Id{Id: category.ID}, ctx)
}

func (catApi *CategoryApi) CreateOneFather(ctx *gin.Context) {
	var requestData request.CategoryCreateOneFather
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	var account accountModel.Account
	err := account.SelectById(requestData.AccountId)
	if responseError(err, ctx) {
		return
	}
	var father categoryModel.Father
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			father, err = categoryService.CreateOneFather(account, requestData.IncomeExpense, requestData.Name, tx)
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.Id{Id: father.ID}, ctx)
}

func (catApi *CategoryApi) MoveCategory(ctx *gin.Context) {
	var requestData request.CategoryMoveCategory
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	category, pass := contextFunc.GetCategoryByParam(ctx)
	if !pass {
		return
	}
	txFunc := func(tx *gorm.DB) error {
		user, err := contextFunc.GetUser(ctx)
		if err != nil {
			return err
		}
		var previous *categoryModel.Category
		if requestData.Previous != nil {
			previous = &categoryModel.Category{}
			err = global.GvaDb.First(previous, requestData.Previous).Error
			if err != nil {
				return err
			}
		}

		var father *categoryModel.Father
		if requestData.FatherId != nil {
			father = &categoryModel.Father{}
			err = global.GvaDb.First(father, requestData.FatherId).Error
			if err != nil {
				return err
			}
		}
		return categoryService.MoveCategory(category, previous, father, user, tx)
	}
	err := global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) MoveFather(ctx *gin.Context) {
	var requestData request.CategoryMoveFather
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	txFunc := func(tx *gorm.DB) error {
		var father categoryModel.Father
		err := global.GvaDb.First(&father, ctx.Param("id")).Error
		if err != nil {
			return err
		}

		var previous *categoryModel.Father
		if requestData.Previous != nil {
			previous = &categoryModel.Father{}
			err = global.GvaDb.First(previous, requestData.Previous).Error
			if err != nil {
				return err
			}
		}
		return categoryService.MoveFather(father, previous, tx)
	}
	err := global.GvaDb.Transaction(txFunc)
	if handelError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) Update(ctx *gin.Context) {
	var requestData request.CategoryUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	var category categoryModel.Category
	err := global.GvaDb.First(&category, ctx.Param("id")).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	txFunc := func(tx *gorm.DB) error {
		return categoryService.Update(
			category, categoryModel.CategoryUpdateData{Name: requestData.Name, Icon: requestData.Icon}, tx,
		)
	}
	if err = global.GvaDb.Transaction(txFunc); responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) UpdateFather(ctx *gin.Context) {
	var requestData request.Name
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	var father categoryModel.Father
	err := global.GvaDb.First(&father, ctx.Param("id")).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	err = categoryService.UpdateFather(father, requestData.Name)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) GetTree(ctx *gin.Context) {
	var requestData request.CategoryGetTree
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, _, pass := checkFunc.AccountBelongAndGet(requestData.AccountId, ctx)
	if false == pass {
		return
	}
	fatherSequence, err := categoryService.GetSequenceFather(account, requestData.IncomeExpense)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	// 响应
	var responseTree response.CategoryTree
	var categoryList []categoryModel.Category
	responseTree.Tree = make([]response.FatherOne, len(fatherSequence), len(fatherSequence))

	for i, father := range fatherSequence {
		categoryList, err = categoryService.GetSequenceCategoryByFather(father)
		if responseError(err, ctx) {
			return
		}
		err = responseTree.Tree[i].SetData(father, categoryList)
		if responseError(err, ctx) {
			return
		}
	}
	response.OkWithData(responseTree, ctx)
}

func (catApi *CategoryApi) Delete(ctx *gin.Context) {
	var category categoryModel.Category
	err := global.GvaDb.First(&category, ctx.Param("id")).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if pass := checkFunc.AccountBelong(category.AccountId, ctx); pass == false {
		return
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return categoryService.Delete(category, tx)
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) DeleteFather(ctx *gin.Context) {
	var father categoryModel.Father
	err := global.GvaDb.First(&father, ctx.Param("id")).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if pass := checkFunc.AccountBelong(father.AccountId, ctx); pass == false {
		return
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			err = categoryService.DeleteFather(father, tx)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) GetList(ctx *gin.Context) {
	var requestData request.CategoryGetList
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, _, pass := checkFunc.AccountBelongAndGet(requestData.AccountId, ctx)
	if false == pass {
		return
	}
	categoryList, err := categoryModel.NewDao().GetListByAccount(account)
	if responseError(err, ctx) {
		return
	}

	var responseData response.CategoryDetailList
	err = responseData.SetData(categoryList)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.CategoryDetail]{List: responseData}, ctx)
}

func (catApi *CategoryApi) MappingCategory(ctx *gin.Context) {
	// 获取数据
	var requestData request.CategoryMapping
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	parentCategory, pass := contextFunc.GetCategoryByParam(ctx)
	if false == pass {
		return
	}
	childCategory, err := categoryModel.NewDao().SelectById(requestData.ChildCategoryId)
	if responseError(err, ctx) {
		return
	}
	// 执行
	var operator userModel.User
	operator, err = contextFunc.GetUser(ctx)
	txFunc := func(tx *gorm.DB) error {
		_, err = categoryService.MappingCategory(parentCategory, childCategory, operator, tx)
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
	} else if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) DeleteCategoryMapping(ctx *gin.Context) {
	// 获取数据
	var requestData request.CategoryMapping
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	parentCategory, pass := contextFunc.GetCategoryByParam(ctx)
	if false == pass {
		return
	}
	childCategory, err := categoryModel.NewDao().SelectById(requestData.ChildCategoryId)
	if responseError(err, ctx) {
		return
	}
	// 执行
	var operator userModel.User
	operator, err = contextFunc.GetUser(ctx)
	txFunc := func(tx *gorm.DB) error {
		err = categoryService.DeleteMapping(parentCategory, childCategory, operator, tx)
		return err
	}
	err = global.GvaDb.Transaction(txFunc)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) GetMappingTree(ctx *gin.Context) {
	var requestData request.CategoryGetMappingTree
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	parentAccountId, childAccountId := requestData.ParentAccountId, requestData.ChildAccountId
	if !checkFunc.AccountBelong(parentAccountId, ctx) || !checkFunc.AccountBelong(childAccountId, ctx) {
		return
	}

	list, err := categoryModel.NewDao().GetMappingByAccountMappingOrderByChildCategoryWeight(
		parentAccountId, childAccountId,
	)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var responseData response.CategoryMappingTree
	responseData.Tree = make([]response.CategoryMappingTreeFather, 0)
	var lastParentCategoryId uint
	startIndex := 0
	for i, mapping := range list {
		if lastParentCategoryId != mapping.ParentCategoryId {
			if lastParentCategoryId != 0 {
				var responseParent response.CategoryMappingTreeFather
				err = responseParent.SetDataFromCategoryMapping(list[startIndex:i])
				if responseError(err, ctx) {
					return
				}
				responseData.Tree = append(responseData.Tree, responseParent)
				startIndex = i
			}
			lastParentCategoryId = mapping.ParentCategoryId
		}
	}

	if len(list) > 0 {
		var responseParent response.CategoryMappingTreeFather
		err = responseParent.SetDataFromCategoryMapping(list[startIndex:])
		if responseError(err, ctx) {
			return
		}
		responseData.Tree = append(responseData.Tree, responseParent)
	}
	response.OkWithData(responseData, ctx)
}
