package util

import (
	"errors"
	"strings"
)

type dataFunc interface {
	utilDataFunc()
	CopyNotEmptyStringOptional(originData *string, targetData *string) error
}

type data struct{}

var Data data

func (d *data) utilDataFunc() {}

var (
	ErrDataIsEmpty = errors.New("数据不可为空")
)

func (d *data) NotEmptyString(originData *string) error {
	if originData != nil {
		*originData = strings.TrimSpace(*originData)
		if *originData == "" {
			return ErrDataIsEmpty
		}
	}
	return ErrDataIsEmpty
}

func (d *data) CopyNotEmptyStringOptional(originData *string, targetData *string) error {
	if originData != nil {
		*originData = strings.TrimSpace(*originData)
		if *originData == "" {
			return ErrDataIsEmpty
		}
		*targetData = *originData
	}
	return nil
}
