package thirdparty

import (
	_ "KeepAccount/test/initialize"
)
import (
	_commonService "KeepAccount/service/common"
	"testing"
)

var commonServer = _commonService.GroupApp

func TestJwt(t *testing.T) {
	claims := commonServer.MakeCustomClaims(1)
	token, err := commonServer.GenerateJWT(claims)
	if err != nil {
		t.Error(err)
	}
	claims, err = commonServer.ParseToken(token)
	if err != nil {
		t.Error(err)
	}

	t.Log(claims)
}
