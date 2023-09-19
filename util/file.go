package util

import (
	"KeepAccount/global/constant"
	"encoding/csv"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"mime/multipart"
	"path"
)

type file struct{}

var File file

func (f *file) GetFileReader(file *multipart.FileHeader, encoding constant.Encoding) io.Reader {
	uploadedFile, err := file.Open()
	if err != nil {
		return nil
	}
	defer uploadedFile.Close()
	var reader io.Reader
	if encoding == constant.GBK {
		reader = transform.NewReader(uploadedFile, simplifiedchinese.GBK.NewDecoder())
	} else {
		reader = uploadedFile
	}
	return reader
}

func (f *file) GetFileSuffix(filename string) string {
	return path.Ext(filename)
}

func (f *file) GetContentFormCSVReader(reader io.Reader) ([][]string, error) {
	csvReader := csv.NewReader(reader)
	result := [][]string{}
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return result, err
		}
		result = append(result, row)
	}
	return result, nil
}

func (f *file) GetContentFormEXCELReader(reader io.Reader) ([][]string, error) {
	result := [][]string{}
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return result, err
	}
	result, err = file.GetRows(file.GetSheetName(1))
	return result, nil
}
