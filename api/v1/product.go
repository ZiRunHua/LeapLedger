package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	"KeepAccount/model/common/query"
	productModel "KeepAccount/model/product"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type _productApi interface {
	GetList(ctx *gin.Context)
	GetTransactionCategory(ctx *gin.Context)
	MappingTransactionCategory(ctx *gin.Context)
	DeleteTransactionCategoryMapping(ctx *gin.Context)
	GetMappingTree(ctx *gin.Context)
	ImportProductBill(ctx *gin.Context)
}

type ProductApi struct {
}

func (p *ProductApi) GetList(ctx *gin.Context) {
	var product productModel.Product
	rows, err := global.GvaDb.Model(&product).Where("hide = ?", 0).Order("weight desc").Rows()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}

	var responseData response.ProductGetList
	responseData.List = []response.ProductGetOne{}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &product)
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		responseData.List = append(
			responseData.List, response.ProductGetOne{Name: product.Name, UniqueKey: product.Key},
		)
	}
	response.OkWithData(responseData, ctx)
}

func (p *ProductApi) GetTransactionCategory(ctx *gin.Context) {
	var transactionCategory productModel.TransactionCategory
	rows, err := global.GvaDb.Model(&transactionCategory).Where(
		"product_key = ?", ctx.Param("key"),
	).Order("income_expense DESC,weight DESC").Rows()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}

	var responseData response.ProductGetTransactionCategoryList
	responseData.List = []response.ProductGetTransactionCategory{}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &transactionCategory)
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		responseData.List = append(
			responseData.List,
			response.ProductGetTransactionCategory{
				Id: transactionCategory.ID, Name: transactionCategory.Name,
				IncomeExpense: transactionCategory.IncomeExpense,
			},
		)
	}
	response.OkWithData(responseData, ctx)
}

func (p *ProductApi) MappingTransactionCategory(ctx *gin.Context) {
	var transactionCategory productModel.TransactionCategory
	err := global.GvaDb.Model(&transactionCategory).First(&transactionCategory, ctx.Param("id")).Error
	if responseError(err, ctx) {
		return
	}
	var requestData request.ProductMappingTransactionCategory
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	var category categoryModel.Category
	err = global.GvaDb.Model(&category).First(&category, requestData.CategoryId).Error
	if responseError(err, ctx) {
		return
	}
	_, err = productService.MappingTransactionCategory(&category, &transactionCategory)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (p *ProductApi) DeleteTransactionCategoryMapping(ctx *gin.Context) {
	var ptc productModel.TransactionCategory
	err := global.GvaDb.Model(&ptc).First(&ptc, ctx.Param("id")).Error
	if responseError(err, ctx) {
		return
	}
	var requestData request.ProductMappingTransactionCategory
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	pass, category, _ := checkFunc.TransactionCategoryBelong(requestData.CategoryId, ctx)
	if pass == false {
		return
	}

	err = productService.DeleteMappingTransactionCategory(category, &ptc)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (p *ProductApi) GetMappingTree(ctx *gin.Context) {
	var requestData request.ProductGetMappingTree
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); pass == false {
		return
	}
	var prodTransCategory productModel.TransactionCategory
	transCategoryMap, err := prodTransCategory.GetMap(requestData.ProductKey)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var prodTransCategoryIds []uint
	for id := range transCategoryMap {
		prodTransCategoryIds = append(prodTransCategoryIds, id)
	}
	var prodTransCategoryMapping productModel.TransactionCategoryMapping
	rows, err := global.GvaDb.Model(&productModel.TransactionCategoryMapping{}).Preload("TransCategory").Where(
		"account_id = ? AND product_key = ?", requestData.AccountId, requestData.ProductKey,
	).Order("category_id desc").Rows()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}

	var tree response.ProductGetMappingTree
	children := make(map[uint][]uint)
	fatherList := []uint{}
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, &prodTransCategoryMapping)
		if err != nil {
			response.FailToError(ctx, err)
			return
		}

		if children[prodTransCategoryMapping.CategoryID] == nil {
			fatherList = append(fatherList, prodTransCategoryMapping.CategoryID)
		}
		prodTransCategory = transCategoryMap[prodTransCategoryMapping.PtcID]
		children[prodTransCategoryMapping.CategoryID] = append(
			children[prodTransCategoryMapping.CategoryID],
			prodTransCategoryMapping.PtcID,
		)
	}

	for _, fatherId := range fatherList {
		tree.Tree = append(
			tree.Tree, response.ProductGetMappingTreeFather{FatherId: fatherId, Children: children[fatherId]},
		)
	}
	response.OkWithData(tree, ctx)
}

func (p *ProductApi) ImportProductBill(ctx *gin.Context) {
	product, err := p.getProductByParam(ctx)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	file, err := ctx.FormFile("File")
	if err != nil {
		response.FailToError(ctx, err)
		return
	}

	account, err := query.FirstByPrimaryKey[*accountModel.Account](ctx.Param("AccountId"))
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return productService.BillImport.BillImport(account, product, file, tx)
		},
	)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (p *ProductApi) getProductByParam(ctx *gin.Context) (*productModel.Product, error) {
	product := productModel.Product{}
	return product.SelectByPrimaryKey(ctx.Param("key"))
}
