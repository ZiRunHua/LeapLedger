package commonService

import (
	"KeepAccount/global"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"github.com/gin-gonic/gin"
)

type current struct{}

var Current = new(current)

func (c *current) GetClaims(ctx *gin.Context) util.CustomClaims {
	return ctx.MustGet("claims").(util.CustomClaims)
}

func (c *current) GetUserId(ctx *gin.Context) uint {
	return ctx.MustGet("userId").(uint)
}

func (c *current) GetUser(ctx *gin.Context) (*userModel.User, error) {
	user := new(userModel.User)
	err := user.SelectById(c.GetUserId(ctx))
	return user, err
}

func (c *current) GetClient(ctx *gin.Context) global.Client {
	userAgent := ctx.GetHeader("User-Agent")
	for clientType, client := range userModel.GetClients() {
		if client.CheckUserAgent(userAgent) {
			return clientType
		}
	}
	panic("Not found client")
}
