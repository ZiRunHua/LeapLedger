package util

import (
	"KeepAccount/global/constant"
	"encoding/csv"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"mime/multipart"
	"os"
	"path"
)

type file struct{}

var File file

type FileWithSuffix struct {
	Filename string
	Suffix   string
	reader   io.ReadCloser
}

func (f *file) GetNewFileWithSuffixByFileHeader(fileHeader *multipart.FileHeader) (*FileWithSuffix, error) {
	reader, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	return &FileWithSuffix{
		Filename: fileHeader.Filename,
		Suffix:   f.GetFileSuffix(fileHeader.Filename),
		reader:   reader,
	}, nil
}

func (f *file) GetNewFileWithSuffixByFilePath(filePath string) (*FileWithSuffix, error) {
	osFile, err := os.Open(constant.WORK_PATH + filePath)
	if err != nil {
		return nil, err
	}
	name := osFile.Name()
	return &FileWithSuffix{
		Filename: name,
		Suffix:   f.GetFileSuffix(name),
		reader:   osFile,
	}, nil
}

func (fws *FileWithSuffix) GetReaderByEncoding(encoding constant.Encoding) io.Reader {
	return File.GetReaderByEncoding(fws.reader, encoding)
}

func (fws *FileWithSuffix) Close() error {
	return fws.reader.Close()
}

func (f *file) GetReaderByEncoding(reader io.Reader, encoding constant.Encoding) io.Reader {
	if encoding == constant.GBK {
		return transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
	}
	return reader
}

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

type RowHandleFunc func(row []string, err error) (isContinue bool)
type IteratorsHandleReaderFunc func(reader io.Reader, handleFunc RowHandleFunc) error

// 迭代器处理CSV 会跳过空行
func (f *file) IteratorsHandleCSVReader(reader io.Reader, handleFunc RowHandleFunc) error {
	csvReader := csv.NewReader(reader)
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if false == handleFunc(row, err) {
			break
		}
	}
	return nil
}

// 迭代器处理EXCEL 会跳过空行
func (f *file) IteratorsHandleEXCELReader(reader io.Reader, handleFunc RowHandleFunc) error {
	var err error
	file, rows := &excelize.File{}, &excelize.Rows{}
	if file, err = excelize.OpenReader(reader); err != nil {
		return err
	}
	if rows, err = file.Rows(file.GetSheetName(1)); err != nil {
		return err
	}
	for rows.Next() {
		row, err := rows.Columns()
		if len(row) == 0 {
			continue
		}
		if false == handleFunc(row, err) {
			break
		}
	}
	return rows.Close()
}
