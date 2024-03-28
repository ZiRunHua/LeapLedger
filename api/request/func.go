package request

import (
	"KeepAccount/api/response"
	"KeepAccount/global/constant"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"time"
)

func GetJsonRequest(obj any, ctx *gin.Context) {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		response.FailToParameter(ctx, err)
	}
}

//func GetRequest(requestData *any, ctx *gin.Context) {
//	if err := ctx.ShouldBind(requestData); err == nil {
//		return
//	}
//	response.FailToParameter(ctx, err)
//}

func ReaderExcelFile(ctx *gin.Context, encoding constant.Encoding) *csv.Reader {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil
	}

	fmt.Println(file.Header.Get("Content-Type"))
	// 打开上传的文件
	uploadedFile, err := file.Open()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	defer uploadedFile.Close()
	var reader io.Reader
	if encoding == constant.GBK {
		reader = transform.NewReader(uploadedFile, simplifiedchinese.GBK.NewDecoder())
	} else {
		reader = uploadedFile
	}

	return csv.NewReader(reader)
}

func GetTimeByTimestamp(timestamp *int64) *time.Time {
	if timestamp != nil {
		result := time.Unix(*timestamp, 0)
		return &result
	}
	panic("GetTimeByTimestamp error")
}
