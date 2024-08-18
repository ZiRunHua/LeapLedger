#!/bin/sh

echo go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/swaggo/swag/cmd/swag@latest
echo go run /go/LeapLedger/docs/beforeDocsMake/renameModel/main.go
go run /go/LeapLedger/docs/beforeDocsMake/renameModel/main.go
echo swag init -p pascalcase
swag init -p pascalcase