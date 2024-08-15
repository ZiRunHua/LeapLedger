#!/bin/sh
go install github.com/swaggo/swag/cmd/swag@latest

swag init

go run /go/LeapLedger/docs/main/capitalize_properties.go

rm /go/LeapLedger/docs/swagger.yaml