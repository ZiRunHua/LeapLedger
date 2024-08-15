package main

import (
	"KeepAccount/global/constant"
	"encoding/json"
	"fmt"
	"os"
	"unicode"
)

func main() {
	// 读取swagger生成的json文件
	data, err := os.ReadFile(constant.WORK_PATH + "/docs/swagger.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 解析json数据
	var swagger map[string]interface{}
	if err := json.Unmarshal(data, &swagger); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// 获取definitions部分
	definitions := swagger["definitions"].(map[string]interface{})

	for _, def := range definitions {
		defMap := def.(map[string]interface{})
		if defMap["properties"] == nil {
			continue
		}
		properties := defMap["properties"].(map[string]interface{})

		newProperties := make(map[string]interface{})

		for key, value := range properties {
			newKey := CapitalizeFirstLetter(key)
			newProperties[newKey] = value
		}

		defMap["properties"] = newProperties
	}

	// 将修改后的json数据写回文件
	modifiedData, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	if err := os.WriteFile(constant.WORK_PATH+"/docs/swagger.json", modifiedData, 0644); err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("Swagger JSON has been modified successfully.")
}
func CapitalizeFirstLetter(str string) string {
	if len(str) == 0 {
		return str
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
