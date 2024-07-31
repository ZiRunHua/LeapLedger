FROM golang:1.20.8-alpine

# 设置环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io,direct
ENV TZ=Asia/Shanghai

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

CMD ["./leapledger"]

