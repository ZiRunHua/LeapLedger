package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	"KeepAccount/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type ProductApi struct {
}

// GetList
//
//	@Tags		Product
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.ProductList}
//	@Header		200	{string}	Cache-Control	"max-age=604800"
//	@Router		/product/list [get]
func (p *ProductApi) GetList(ctx *gin.Context) {
	var product productModel.Product
	rows, err := db.Db.Model(&product).Where("hide = ?", 0).Rows()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}

	var responseData response.ProductList
	responseData.List = []response.ProductOne{}
	for rows.Next() {
		err = db.Db.ScanRows(rows, &product)
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		responseData.List = append(
			responseData.List, response.ProductOne{Name: product.Name, UniqueKey: string(product.Key)},
		)
	}
	ctx.Header("Cache-Control", "max-age=604800")
	response.OkWithData(responseData, ctx)
}

// GetTransactionCategory
//
//	@Tags		Product/TransCategory
//	@Produce	json
//	@Param		key	path		int	true	"Product unique key"
//	@Success	200	{object}	response.Data{Data=response.ProductTransactionCategoryList}
//	@Header		200	{string}	Cache-Control	"max-age=604800"
//	@Router		/product/{key}/transCategory [get]
func (p *ProductApi) GetTransactionCategory(ctx *gin.Context) {
	var transactionCategory productModel.TransactionCategory
	rows, err := db.Db.Model(&transactionCategory).Where(
		"product_key = ?", ctx.Param("key"),
	).Order("income_expense DESC").Rows()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}

	var responseData response.ProductTransactionCategoryList
	responseData.List = []response.ProductTransactionCategory{}
	for rows.Next() {
		err = db.Db.ScanRows(rows, &transactionCategory)
		if err != nil {
			response.FailToError(ctx, err)
			return
		}
		responseData.List = append(
			responseData.List,
			response.ProductTransactionCategory{
				Id: transactionCategory.ID, Name: transactionCategory.Name,
				IncomeExpense: transactionCategory.IncomeExpense,
			},
		)
	}
	ctx.Header("Cache-Control", "max-age=604800")
	response.OkWithData(responseData, ctx)
}

// MappingTransactionCategory
//
//	@Tags		Product/TransCategory/Mapping
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int											true	"Account ID"
//	@Param		id			path		int											true	"Product transaction category ID"
//	@Param		body		body		request.ProductMappingTransactionCategory	true	"data"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/product/transCategory/{id}/mapping [post]
func (p *ProductApi) MappingTransactionCategory(ctx *gin.Context) {
	var transactionCategory productModel.TransactionCategory
	err := db.Db.Model(&transactionCategory).First(&transactionCategory, ctx.Param("id")).Error
	if responseError(err, ctx) {
		return
	}
	var requestData request.ProductMappingTransactionCategory
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	var category categoryModel.Category
	err = db.Db.Model(&category).First(&category, requestData.CategoryId).Error
	if responseError(err, ctx) {
		return
	}
	if category.AccountId != contextFunc.GetAccountId(ctx) {
		response.FailToParameter(ctx, global.ErrAccountId)
		return
	}
	err = db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			_, err = productService.MappingTransactionCategory(category, transactionCategory, ctx)
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// DeleteTransactionCategoryMapping
//
//	@Tags		Product/TransCategory/Mapping
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int											true	"Account ID"
//	@Param		id			path		int											true	"Product transaction category ID"
//	@Param		body		body		request.ProductMappingTransactionCategory	true	"data"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/product/transCategory/{id}/mapping [delete]
func (p *ProductApi) DeleteTransactionCategoryMapping(ctx *gin.Context) {
	var ptc productModel.TransactionCategory
	err := db.Db.Model(&ptc).First(&ptc, ctx.Param("id")).Error
	if responseError(err, ctx) {
		return
	}
	var requestData request.ProductMappingTransactionCategory
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	pass, category, _ := checkFunc.TransactionCategoryBelongAndGet(requestData.CategoryId, ctx)
	if pass == false {
		return
	}

	err = db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			return productService.DeleteMappingTransactionCategory(category, ptc, ctx)
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// GetMappingTree
//
//	@Tags		Product/TransCategory/Mapping
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int		true	"Account ID"
//	@Param		key			path		string	true	"Product unique key"
//	@Success	200			{object}	response.Data{Data=response.ProductMappingTree}
//	@Router		/account/{accountId}/product/{key}/transCategory/mapping/tree [get]
func (p *ProductApi) GetMappingTree(ctx *gin.Context) {
	accountId, productKey := contextFunc.GetAccountId(ctx), productModel.KeyValue(ctx.Param("key"))
	var prodTransCategory productModel.TransactionCategory
	transCategoryMap, err := prodTransCategory.GetMap(productKey)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	var prodTransCategoryIds []uint
	for id := range transCategoryMap {
		prodTransCategoryIds = append(prodTransCategoryIds, id)
	}
	var prodTransCategoryMapping productModel.TransactionCategoryMapping
	rows, err := db.Db.Model(&productModel.TransactionCategoryMapping{}).Preload("TransCategory").Where(
		"account_id = ? AND product_key = ?", accountId, productKey,
	).Order("category_id desc").Rows()
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	// response
	var tree response.ProductMappingTree
	children := make(map[uint][]uint)
	var fatherList []uint
	for rows.Next() {
		err = db.Db.ScanRows(rows, &prodTransCategoryMapping)
		if err != nil {
			response.FailToError(ctx, err)
			return
		}

		if children[prodTransCategoryMapping.CategoryId] == nil {
			fatherList = append(fatherList, prodTransCategoryMapping.CategoryId)
		}
		prodTransCategory = transCategoryMap[prodTransCategoryMapping.PtcId]
		children[prodTransCategoryMapping.CategoryId] = append(
			children[prodTransCategoryMapping.CategoryId],
			prodTransCategoryMapping.PtcId,
		)
	}

	for _, fatherId := range fatherList {
		tree.Tree = append(
			tree.Tree, response.ProductMappingTreeFather{FatherId: fatherId, Children: children[fatherId]},
		)
	}
	response.OkWithData(tree, ctx)
}

// ImportProductBill
//
//	@Tags		Product/Bill/Import
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		accountId	path		int		true	"Account ID"
//	@Param		key			path		int		true	"Product unique key"
//	@Param		file		formData	file	true	"file to upload"
//	@Success	200			{object}	response.NoContent
//	@Router		/account/{accountId}/product/{key}/bill/import [post]
func (p *ProductApi) ImportProductBill(conn *websocket.Conn, ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("File")
	if responseError(err, ctx) {
		return
	}
	var file *util.FileWithSuffix
	file, err = util.File.GetNewFileWithSuffixByFileHeader(fileHeader)
	if responseError(err, ctx) {
		return
	}
	defer file.Close()
	account, accountUser := contextFunc.GetAccount(ctx), contextFunc.GetAccountUser(ctx)
	product, err := p.getProductByParam(ctx)
	if responseError(err, ctx) {
		return
	}
	err = db.Db.Transaction(
		func(tx *gorm.DB) error {
			return productService.BillImport(accountUser, account, product, file, tx)
		},
	)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

func (p *ProductApi) getProductByParam(ctx *gin.Context) (productModel.Product, error) {
	product := productModel.Product{}
	return product.SelectByKey(productModel.KeyValue(ctx.Param("key")))
}
