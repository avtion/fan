FROM golang:1.16.6-alpine AS dev
# 安装gops用于DEBUG
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct
RUN go install github.com/google/gops@latest
# 拷贝二进制程序
COPY . .
RUN go mod download && go build -o main .
ENTRYPOINT ["./main"]