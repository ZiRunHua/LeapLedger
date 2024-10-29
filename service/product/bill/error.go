package bill

import "errors"

var (
	ErrCategoryCannotRead      = errors.New("交易类型不可读取")
	ErrCategoryReadFail        = errors.New("读取交易类型失败")
	ErrCategoryMappingNotExist = errors.New("不存在关联类型")
)
