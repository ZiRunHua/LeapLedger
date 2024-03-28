package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
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
	if pass := checkFunc.AccountBelong(father.AccountID, ctx); false == pass {
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
	var category categoryModel.Category
	var previous *categoryModel.Category
	err := global.GvaDb.First(&category, ctx.Param("id")).Error
	if handelError(err, ctx) {
		return
	}
	var requestData request.CategoryMoveCategory
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if requestData.Previous != nil {
		err = global.GvaDb.First(previous, requestData.Previous).Error
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		if previous.ID == category.ID {
			response.FailWithMessage("数据异常", ctx)
			return
		}
	}

	father := categoryModel.Father{}
	if requestData.FatherId != nil {
		err = global.GvaDb.First(&father, requestData.FatherId).Error
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return categoryService.MoveCategory(category, previous, father, tx)
		},
	)

	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) MoveFather(ctx *gin.Context) {
	var father categoryModel.Father
	var previous *categoryModel.Father
	err := global.GvaDb.First(&father, ctx.Param("id")).Error
	if handelError(err, ctx) {
		return
	}
	var requestData request.CategoryMoveFather
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if requestData.Previous != nil {
		err = global.GvaDb.First(previous, requestData.Previous).Error
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		if previous.ID == father.ID {
			response.FailWithMessage("数据异常", ctx)
			return
		}
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return categoryService.MoveFather(father, previous, tx)
		},
	)
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
	categorySequence, err := categoryService.GetSequenceCategory(account, requestData.IncomeExpense)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var responseTree response.CategoryTree
	var responseChildren []response.CategoryOne
	for _, father := range fatherSequence {
		responseChildren = make([]response.CategoryOne, 0)
		if categorySequence[father.ID] != nil {
			for _, category := range *categorySequence[father.ID] {
				responseChildren = append(
					responseChildren,
					*response.CategoryModelToResponse(&category),
				)
			}
		}
		responseTree.Tree = append(
			responseTree.Tree,
			response.FatherOne{
				Name: father.Name, Id: father.ID, IncomeExpense: father.IncomeExpense, Children: responseChildren,
			},
		)
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
	if pass := checkFunc.AccountBelong(father.AccountID, ctx); pass == false {
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
	var requestData request.CategoryGetTree
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	var account accountModel.Account
	err = global.GvaDb.First(&account, requestData.AccountId).Error
	if err != nil {
		response.FailToError(ctx, errors.Wrap(err, ""))
		return
	}
	fatherSequence, err := categoryService.GetSequenceFather(account, requestData.IncomeExpense)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	categorySequence, err := categoryService.GetSequenceCategory(account, requestData.IncomeExpense)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var responseTree response.TwoLevelTree
	var responseChildren []response.NameId
	for _, father := range fatherSequence {
		responseChildren = make([]response.NameId, 0)
		if categorySequence[father.ID] != nil {
			for _, category := range *categorySequence[father.ID] {
				responseChildren = append(responseChildren, response.NameId{Name: category.Name, Id: category.ID})
			}
		}
		responseTree.Tree = append(
			responseTree.Tree,
			response.Father{NameId: response.NameId{Name: father.Name, Id: father.ID}, Children: responseChildren},
		)
	}
	response.OkWithData(responseTree, ctx)
}
