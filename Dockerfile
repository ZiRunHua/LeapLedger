# 使用官方的Go镜像作为基础镜像
FROM golang:1.20-alpine

# 设置环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io,direct
ENV TZ=Asia/Shanghai

# 安装curl
RUN apk add --no-cache curl

## 下载并安装NATS服务器
#RUN curl -sSL https://binaries.nats.dev/nats-io/nats-server/v2@latest -o /usr/local/bin/nats-server \
#    && chmod +x /usr/local/bin/nats-server
## 指定容器启动命令
#CMD ["./nats-server"]
# 设置工作目录
WORKDIR /go/LeapLedger

# 拷贝依赖文件并下载
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# 拷贝源代码
COPY . .

# 构建Go应用程序
RUN go build -o leapledger

# 声明服务端口
EXPOSE 8080

# 指定容器启动命令
CMD ["./leapledger"]
