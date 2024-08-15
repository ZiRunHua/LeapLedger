# Makefile

# 定义变量
SWAGGER_JSON=./docs/swagger.json
MODIFIED_SWAGGER_JSON=./docs/swagger_modified.json

# 生成 swagger 文档
swagger: $(SWAGGER_JSON)

# 生成 swagger.json 文件
$(SWAGGER_JSON):
	@echo "Generating Swagger JSON..."
	swag init

# 修改 swagger.json 的 properties 的 key 的首字母为大写
$(MODIFIED_SWAGGER_JSON): $(SWAGGER_JSON)
	@echo "Capitalizing properties keys..."
	go run ./docs/capitalize_properties.go

# 默认目标，先生成 swagger 文件，然后修改 properties
all: swagger $(MODIFIED_SWAGGER_JSON)

# 清理生成的文件
clean:
	rm -f $(SWAGGER_JSON) $(MODIFIED_SWAGGER_JSON)
