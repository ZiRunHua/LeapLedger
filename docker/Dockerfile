FROM golang:1.23-alpine

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.io,direct
ENV TZ=Asia/Shanghai

WORKDIR /go/LeapLedger

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN go build -o leapledger

EXPOSE 8080

CMD ["/bin/sh", "-c", "./leapledger"]