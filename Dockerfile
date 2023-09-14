# 使用官方的Go镜像作为基础镜像
FROM golang:1.20-alpine

# 设置环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io,direct
ENV TZ=Asia/Shanghai
# 设置工作目录
WORKDIR /server/KeepAccount

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
# 将本地代码复制到容器中
COPY . .

# 构建Go应用程序
RUN go build -o keepaccount

# 声明服务端口
EXPOSE 8080

# 指定容器启动命令
CMD ["./keepaccount"]
