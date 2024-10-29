package v1

import (
	"github.com/ZiRunHua/LeapLedger/api/request"
	"github.com/ZiRunHua/LeapLedger/api/response"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/db"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CategoryApi struct {
}

// CreateOne
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int							true	"Account ID"
//	@Param		body		body		request.CategoryCreateOne	true	"category data"
//	@Success	200			{object}	response.Data{Data=response.CategoryOne}
//	@Router		/account/{accountId}/category [post]
func (catApi *CategoryApi) CreateOne(ctx *gin.Context) {
	var requestData request.CategoryCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	var father categoryModel.Father
	err := father.SelectById(requestData.FatherId)
	if responseError(err, ctx) {
		return
	}
	if father.AccountId != contextFunc.GetAccountId(ctx) {
		response.FailToError(ctx, global.ErrAccountId)
		return
	}

	var category categoryModel.Category
	category, err = categoryService.CreateOne(
		father,
		categoryService.NewCategoryData(requestData.Name, requestData.Icon),
		ctx,
	)
	if responseError(err, ctx) {
		return
	}

	var responseData response.CategoryOne
	err = responseData.SetData(category)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// CreateOneFather
//
//	@Tags		Category/Father
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int								true	"Account ID"
//	@Param		body		body		request.CategoryCreateOneFather	true	"father category data"
//	@Success	200			{object}	response.Data{Data=response.FatherOne}
//	@Router		/account/{accountId}/category/father [post]
func (catApi *CategoryApi) CreateOneFather(ctx *gin.Context) {
	var requestData request.CategoryCreateOneFather
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	father, err := categoryService.CreateOneFather(
		contextFunc.GetAccount(ctx), requestData.IncomeExpense, requestData.Name, ctx,
	)
	if responseError(err, ctx) {
		return
	}

	var responseData response.FatherOne
	err = responseData.SetData(father, []categoryModel.Category{})
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// MoveCategory
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int						true	"Account ID"
//	@Param		id			path		int						true	"Category ID"
//	@Param		body		body		request.CategoryMove	true	"move data"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/category/{id}/move [put]
func (catApi *CategoryApi) MoveCategory(ctx *gin.Context) {
	var requestData request.CategoryMove
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	category, pass := contextFunc.GetCategoryByParam(ctx)
	if !pass {
		return
	}
	var (
		pPrevious *categoryModel.Category
		pFather   *categoryModel.Father
	)
	if requestData.Previous != nil {
		previous, pass := checkFunc.CategoryBelongAndGet(*requestData.Previous, contextFunc.GetAccountId(ctx), ctx)
		if !pass {
			return
		}
		pPrevious = &previous
	}
	if requestData.FatherId != nil {
		father, pass := checkFunc.CategoryFatherBelongAndGet(*requestData.FatherId, contextFunc.GetAccountId(ctx), ctx)
		if !pass {
			return
		}
		pFather = &father
	}
	user, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	err = categoryService.MoveCategory(category, pPrevious, pFather, user, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// MoveFather
//
//	@Tags		Category/Father
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int							true	"Account ID"
//	@Param		id			path		int							true	"Father ID"
//	@Param		body		body		request.CategoryMoveFather	true	"move data"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/category/father/{id}/move [put]
func (catApi *CategoryApi) MoveFather(ctx *gin.Context) {
	var requestData request.CategoryMoveFather
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	var (
		father    categoryModel.Father
		pPrevious *categoryModel.Father
	)
	father, pass := checkFunc.CategoryFatherBelongAndGet(contextFunc.GetId(ctx), contextFunc.GetAccountId(ctx), ctx)
	if !pass {
		return
	}
	if requestData.Previous != nil {
		previous, pass := checkFunc.CategoryFatherBelongAndGet(
			*requestData.Previous, contextFunc.GetAccountId(ctx), ctx,
		)
		if !pass {
			return
		}
		pPrevious = &previous
	}
	err := categoryService.MoveFather(father, pPrevious, ctx)
	if handelError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// Update
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int							true	"Account ID"
//	@Param		id			path		int							true	"Category ID"
//	@Param		body		body		request.CategoryUpdateOne	true	"update data"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/category/{id} [put]
func (catApi *CategoryApi) Update(ctx *gin.Context) {
	var requestData request.CategoryUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	category, pass := contextFunc.GetCategoryByParam(ctx)
	if !pass {
		return
	}
	var err error
	category, err = categoryService.Update(
		contextFunc.GetId(ctx), categoryModel.CategoryUpdateData{Name: requestData.Name, Icon: requestData.Icon}, ctx,
	)
	if responseError(err, ctx) {
		return
	}
	var responseData response.CategoryOne
	err = responseData.SetData(category)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// UpdateFather
//
//	@Tags		Category/Father
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int				true	"Account ID"
//	@Param		id			path		int				true	"Father ID"
//	@Param		body		body		request.Name	true	"update data"
//	@Success	200			{object}	response.Data{Data=response.FatherOne}
//	@Router		/account/{accountId}/category/father/{id} [put]
func (catApi *CategoryApi) UpdateFather(ctx *gin.Context) {
	var requestData request.Name
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	father, pass := contextFunc.GetCategoryFatherByParam(ctx)
	if !pass {
		return
	}
	father, err := categoryService.UpdateFather(father, requestData.Name)
	if responseError(err, ctx) {
		return
	}
	var responseData response.FatherOne
	err = responseData.SetData(father, []categoryModel.Category{})
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetTree
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int						true	"Account ID"
//	@Param		body		body		request.CategoryGetTree	true	"query condition"
//	@Success	200			{object}	response.Data{Data=response.CategoryTree}
//	@Router		/account/{accountId}/category/tree [get]
func (catApi *CategoryApi) GetTree(ctx *gin.Context) {
	var requestData request.CategoryGetTree
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account := contextFunc.GetAccount(ctx)
	fatherSequence, err := categoryService.GetSequenceFather(account, requestData.IncomeExpense)
	if responseError(err, ctx) {
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

// GetList
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int						true	"Account ID"
//	@Param		body		body		request.CategoryGetList	true	"query condition"
//	@Success	200			{object}	response.Data{Data=response.List[response.CategoryDetail]{}}
//	@Router		/account/{accountId}/category/tree [get]
func (catApi *CategoryApi) GetList(ctx *gin.Context) {
	var requestData request.CategoryGetList
	var err error
	if err = ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account := contextFunc.GetAccount(ctx)
	fatherSequence, err := categoryService.GetSequenceFather(account, requestData.IncomeExpense)
	if responseError(err, ctx) {
		return
	}
	// 响应
	var list, categoryList []categoryModel.Category
	var responseData response.CategoryDetailList
	for _, father := range fatherSequence {
		categoryList, err = categoryService.GetSequenceCategoryByFather(father)
		if responseError(err, ctx) {
			return
		}
		list = append(list, categoryList...)
	}
	err = responseData.SetData(list)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.CategoryDetail]{List: responseData}, ctx)
}

// Delete
//
//	@Tags		Category
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Param		id			path		int	true	"Category ID"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/category/{id} [delete]
func (catApi *CategoryApi) Delete(ctx *gin.Context) {
	var category categoryModel.Category
	err := db.Db.First(&category, ctx.Param("id")).Error
	if responseError(err, ctx) {
		return
	}
	if contextFunc.GetAccountId(ctx) != category.AccountId {
		response.FailToParameter(ctx, global.ErrAccountId)
		return
	}
	err = categoryService.Delete(category, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// DeleteFather
//
//	@Tags		Category/Father
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Param		id			path		int	true	"Father ID"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/category/father/{id} [delete]
func (catApi *CategoryApi) DeleteFather(ctx *gin.Context) {
	var father categoryModel.Father
	err := db.Db.First(&father, ctx.Param("id")).Error
	if responseError(err, ctx) {
		return
	}
	err = categoryService.DeleteFather(father, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// MappingCategory
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int						true	"Account ID"
//	@Param		id			path		int						true	"Category ID"
//	@Param		body		body		request.CategoryMapping	true	"data"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/category/{id}/mapping [post]
func (catApi *CategoryApi) MappingCategory(ctx *gin.Context) {
	var requestData request.CategoryMapping
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	parentCategory, pass := contextFunc.GetCategoryByParam(ctx)
	if false == pass {
		return
	}
	childCategory, err := categoryModel.NewDao(db.Get(ctx)).SelectById(requestData.ChildCategoryId)
	if responseError(err, ctx) {
		return
	}
	if !checkFunc.AccountPermission(childCategory.AccountId, accountModel.UserPermissionAddOwn, ctx) {
		return
	}
	// handle
	var operator userModel.User
	operator, err = contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	_, err = categoryService.MappingCategory(parentCategory, childCategory, operator, ctx)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
	} else if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// MappingCategory
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int						true	"Account ID"
//	@Param		id			path		int						true	"Category ID"
//	@Param		body		body		request.CategoryMapping	true	"data"
//	@Success	204			{object}	response.NoContent
//	@Router		/account/{accountId}/category/{id}/mapping [delete]
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
	childCategory, err := categoryModel.NewDao(db.Get(ctx)).SelectById(requestData.ChildCategoryId)
	if responseError(err, ctx) {
		return
	}
	if !checkFunc.AccountPermission(childCategory.AccountId, accountModel.UserPermissionAddOwn, ctx) {
		return
	}
	// 执行
	var operator userModel.User
	operator, err = contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}
	err = categoryService.DeleteMapping(parentCategory, childCategory, operator, ctx)
	if responseError(err, ctx) {
		return
	}
	response.Ok(ctx)
}

// GetMappingTree
//
//	@Tags		Category
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int								true	"Account ID"
//	@Param		body		body		request.CategoryGetMappingTree	true	"query condition"
//	@Success	200			{object}	response.Data{Data=response.CategoryMappingTree}
//	@Router		/account/{accountId}/category/mapping/tree [get]
func (catApi *CategoryApi) GetMappingTree(ctx *gin.Context) {
	var requestData request.CategoryGetMappingTree
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if !checkFunc.AccountBelong(requestData.MappingAccountId, ctx) {
		return
	}

	list, err := categoryModel.NewDao().GetMappingByAccountMappingOrderByParentCategory(
		contextFunc.GetAccountId(ctx), requestData.MappingAccountId,
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
