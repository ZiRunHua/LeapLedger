package response

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	categoryModel "KeepAccount/model/category"
	"KeepAccount/util/dataTool"
	"github.com/pkg/errors"
)

type CategoryOne struct {
	Id            uint
	Name          string
	AccountId     uint
	Icon          string
	IncomeExpense constant.IncomeExpense
}

func (co *CategoryOne) SetData(category categoryModel.Category) error {
	co.Id = category.ID
	co.Name = category.Name
	co.AccountId = category.AccountId
	co.Icon = category.Icon
	co.IncomeExpense = category.IncomeExpense
	return nil
}

type CategoryDetail struct {
	Id            uint
	Name          string
	Icon          string
	FatherId      uint
	FatherName    string
	IncomeExpense constant.IncomeExpense
}

func (cd *CategoryDetail) SetData(category categoryModel.Category, father categoryModel.Father) error {
	cd.Id = category.ID
	cd.Name = category.Name
	cd.Icon = category.Icon
	cd.FatherId = father.ID
	cd.FatherName = father.Name
	cd.IncomeExpense = category.IncomeExpense
	return nil
}

type CategoryDetailList []CategoryDetail

func (cdl *CategoryDetailList) SetData(categoryList dataTool.Slice[uint, categoryModel.Category]) error {
	*cdl = make(CategoryDetailList, len(categoryList), len(categoryList))
	if len(categoryList) == 0 {
		return nil
	}

	fatherIds := categoryList.ExtractValues(func(category categoryModel.Category) uint { return category.FatherId })
	var fatherList dataTool.Slice[uint, categoryModel.Father]
	err := db.Db.Where("id IN (?)", fatherIds).Find(&fatherList).Error
	if err != nil {
		return err
	}
	fatherMap := fatherList.ToMap(func(father categoryModel.Father) uint { return father.ID })

	for i, category := range categoryList {
		err = (*cdl)[i].SetData(category, fatherMap[category.FatherId])
		if err != nil {
			return err
		}
	}
	return nil
}

type FatherOne struct {
	Id            uint
	Name          string
	AccountId     uint
	IncomeExpense constant.IncomeExpense
	Children      []CategoryOne
}

func (fo *FatherOne) SetData(father categoryModel.Father, categoryList []categoryModel.Category) error {
	fo.Id = father.ID
	fo.Name = father.Name
	fo.AccountId = father.AccountId
	fo.IncomeExpense = father.IncomeExpense
	fo.Children = make([]CategoryOne, len(categoryList), len(categoryList))
	var err error
	for i, category := range categoryList {
		err = fo.Children[i].SetData(category)
		if err != nil {
			return err
		}
	}
	return nil
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
