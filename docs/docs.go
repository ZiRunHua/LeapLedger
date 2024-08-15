// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/transaction/{id}": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/response.Data"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Data": {
                                            "$ref": "#/definitions/response.TransactionDetail"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "constant.IncomeExpense": {
            "type": "string",
            "enum": [
                "income",
                "expense"
            ],
            "x-enum-varnames": [
                "Income",
                "Expense"
            ]
        },
        "response.Data": {
            "type": "object",
            "properties": {
                "data": {},
                "msg": {
                    "type": "string"
                }
            }
        },
        "response.TransactionDetail": {
            "type": "object",
            "properties": {
                "accountId": {
                    "type": "integer"
                },
                "accountName": {
                    "type": "string"
                },
                "amount": {
                    "type": "integer"
                },
                "categoryFatherName": {
                    "type": "string"
                },
                "categoryIcon": {
                    "type": "string"
                },
                "categoryId": {
                    "type": "integer"
                },
                "categoryName": {
                    "type": "string"
                },
                "createTime": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "incomeExpense": {
                    "$ref": "#/definitions/constant.IncomeExpense"
                },
                "remark": {
                    "type": "string"
                },
                "tradeTime": {
                    "type": "string"
                },
                "updateTime": {
                    "type": "string"
                },
                "userId": {
                    "type": "integer"
                },
                "userName": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
