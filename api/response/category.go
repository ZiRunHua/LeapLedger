package response

import (
	"KeepAccount/global/constant"
	categoryModel "KeepAccount/model/category"
	"github.com/pkg/errors"
)

func CategoryModelToResponse(category *categoryModel.Category) *CategoryOne {
	if category == nil {
		return &CategoryOne{}
	}
	return &CategoryOne{
		Id:            category.ID,
		Name:          category.Name,
		Icon:          category.Icon,
		IncomeExpense: category.IncomeExpense,
	}
}

type CategoryOne struct {
	Id            uint
	Name          string
	Icon          string
	IncomeExpense constant.IncomeExpense
}

func (co *CategoryOne) SetData(category categoryModel.Category) error {
	co.Id = category.ID
	co.Name = category.Name
	co.Icon = category.Icon
	co.IncomeExpense = category.IncomeExpense
	return nil
}

type FatherOne struct {
	Id            uint
	Name          string
	IncomeExpense constant.IncomeExpense
	Children      []CategoryOne
}

type CategoryTree struct {
	Tree []FatherOne
}

type CategoryMappingTree struct {
	Tree []CategoryMappingTreeFather
}

type CategoryMappingTreeFather struct {
	FatherId    uint
	ChildrenIds []uint
}

func (m *CategoryMappingTreeFather) SetDataFromCategoryMapping(data []categoryModel.Mapping) error {
	if len(data) == 0 {
		return errors.New("data len error")
	}
	m.FatherId = data[0].ParentCategoryId
	m.ChildrenIds = make([]uint, len(data), len(data))
	for i, mapping := range data {
		if m.FatherId != mapping.ParentCategoryId {
			return errors.New("err ParentCategoryId")
		}
		m.ChildrenIds[i] = mapping.ChildCategoryId
	}
	return nil
}
