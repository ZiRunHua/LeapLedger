# keepAccount

## server项目

```bash

# 克隆项目
git clone https://github.com/ZiRunHua/KeepAccount.git

# 构建Docker镜像
docker build -t keepaccount-user .

# 运行容器映射8080端口 
docker run --name keepaccount-user-server -d -p 8080:8080 keepaccount-user

```