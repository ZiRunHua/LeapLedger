package request

import (
	"KeepAccount/api/response"
	"KeepAccount/global"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"net/http"
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

func ReaderExcelFile(ctx *gin.Context, encoding global.Encoding) *csv.Reader {
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
	if encoding == global.GBK {
		reader = transform.NewReader(uploadedFile, simplifiedchinese.GBK.NewDecoder())
	} else {
		reader = uploadedFile
	}

	return csv.NewReader(reader)
}
