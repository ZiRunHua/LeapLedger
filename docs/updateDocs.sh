#!/bin/sh
echo go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/swaggo/swag/cmd/swag@latest
#echo go run /go/LeapLedger/docs/beforeDocsMake/renameModel/main.go
#go run /go/LeapLedger/docs/beforeDocsMake/renameModel/main.go
echo swag init -p pascalcase
swag init -p pascalcase

echo apk add graphviz
apk add graphviz

echo go run /go/LeapLedger/docs/makeVectors/main.go
go run /go/LeapLedger/docs/makeVectors/main.go

echo dot -Tsvg /go/LeapLedger/docs/eventVectorGraph.dot -o /go/LeapLedger/docs/eventVectorGraph.svg
dot -Tsvg /go/LeapLedger/docs/eventVectorGraph.dot -o /go/LeapLedger/docs/eventVectorGraph.svg