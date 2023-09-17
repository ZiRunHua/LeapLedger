package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	"KeepAccount/model/common/query"
	"KeepAccount/util"
	"fmt"
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
	father := new(categoryModel.Father)
	err := father.SelectById(requestData.FatherId, false)
	if pass, _ := checkAccountBelong(father.AccountID, ctx); false == pass {
		return
	}
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	category, err := categoryService.CreateOne(father, requestData.Name)
	if err != nil {
		response.FailToError(ctx, err)
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
	account, err := query.FirstByPrimaryKey[*accountModel.Account](requestData.AccountId)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	father, err := categoryService.CreateOneFather(account, requestData.IncomeExpense, requestData.Name)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.OkWithData(response.Id{Id: father.ID}, ctx)
}

func (catApi *CategoryApi) MoveCategory(ctx *gin.Context) {
	var category, previous categoryModel.Category
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
		err = global.GvaDb.First(&previous, requestData.Previous).Error
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
			util.SetTxOfModels(tx, &category, &previous)
			return categoryService.MoveCategory(&category, &previous, &father)
		},
	)

	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) MoveFather(ctx *gin.Context) {
	var previous, father categoryModel.Father
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
		err = global.GvaDb.First(&previous, requestData.Previous).Error
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		if previous.ID == father.ID {
			response.FailWithMessage("数据异常", ctx)
			return
		}
	}
	err = categoryService.MoveFather(&father, &previous)
	if handelError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (catApi *CategoryApi) Update(ctx *gin.Context) {
	var requestData request.Name
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
	err = categoryService.Update(&category, requestData.Name)
	if err != nil {
		response.FailToError(ctx, err)
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
	err = categoryService.UpdateFather(&father, requestData.Name)
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
	var account accountModel.Account
	err = global.GvaDb.First(&account, requestData.AccountId).Error
	if err != nil {
		response.FailToError(ctx, errors.Wrap(err, ""))
		return
	}
	fatherSequence, err := categoryService.GetSequenceFather(&account, requestData.IncomeExpense)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	fmt.Println(fatherSequence)
	categorySequence, err := categoryService.GetSequenceCategory(&account, requestData.IncomeExpense)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	fmt.Println(categorySequence)
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

func (catApi *CategoryApi) Delete(ctx *gin.Context) {
	var category categoryModel.Category
	err := global.GvaDb.First(&category, ctx.Param("id")).Error
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	if pass, _ := checkAccountBelong(category.AccountID, ctx); pass == false {
		return
	}
	err = categoryService.Delete(&category)
	if err != nil {
		response.FailToError(ctx, err)
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
	if pass, _ := checkAccountBelong(father.AccountID, ctx); pass == false {
		return
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			father.SetTx(tx)
			err = categoryService.DeleteFather(&father)
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
	fatherSequence, err := categoryService.GetSequenceFather(&account, requestData.IncomeExpense)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	categorySequence, err := categoryService.GetSequenceCategory(&account, requestData.IncomeExpense)
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
