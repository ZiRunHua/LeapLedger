package fileTool

import (
	"KeepAccount/global/constant"
	"encoding/csv"
	"errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
	"io"
	"path"
	"strings"
)

type FileWithSuffix struct {
	Filename string
	Suffix   string
	reader   io.ReadCloser
}

func (fws *FileWithSuffix) GetReaderByEncoding(encoding constant.Encoding) io.Reader {
	return GetReaderByEncoding(fws.reader, encoding)
}

func (fws *FileWithSuffix) Close() error {
	return fws.reader.Close()
}

func GetReaderByEncoding(reader io.Reader, encoding constant.Encoding) io.Reader {
	if encoding == constant.GBK {
		return transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
	}
	return reader
}

func GetFileSuffix(filename string) string {
	return path.Ext(filename)
}

func NewRowReader(reader io.Reader, suffix string) (func(yield func([]string) bool), error) {
	switch suffix {
	case ".csv":
		return IteratorsHandleCSVReader(reader)
	case ".excel":
		return IteratorsHandleEXCELReader(reader)
	default:
		return nil, errors.New("不支持该文件类型")
	}
}

func IteratorsHandleCSVReader(reader io.Reader) (func(yield func([]string) bool), error) {
	return func(yield func([]string) bool) {
		csvReader := csv.NewReader(reader)
		for {
			row, err := csvReader.Read()
			if err != nil || !yield(row) {
				return
			}
		}
	}, nil
}

// 迭代器处理EXCEL 会跳过空行
func IteratorsHandleEXCELReader(reader io.Reader) (func(yield func([]string) bool), error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	rows, err := file.Rows(file.GetSheetName(1))
	if err != nil {
		return nil, err
	}
	return func(yield func([]string) bool) {
		defer rows.Close()
		var row []string
		var err error
		for rows.Next() {
			row, err = rows.Columns()
			if err != nil {
				return
			}
			if len(row) == 0 {
				continue
			}
			if !yield(row) {
				return
			}
		}
	}, nil
}

func ExecSqlFile(reader io.Reader, db *gorm.DB) error {
	sqlBytes, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	sqlStatements := strings.Split(string(sqlBytes), ";")
	for _, stmt := range sqlStatements {
		trimmedStmt := strings.TrimSpace(stmt)
		if len(trimmedStmt) > 0 {
			if err = db.Exec(trimmedStmt).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
