FROM golang:1.16.6-alpine AS dev
WORKDIR /go/src/app
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct
RUN go install github.com/google/gops@latest
COPY . .
RUN go mod download && go build -o main .
ENTRYPOINT ["./main"]
CMD ["-u", "", "-p", "", "-w", ""]