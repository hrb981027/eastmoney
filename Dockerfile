FROM golang:alpine

ENV GO111MODULE=on \
GOPROXY=https://goproxy.io

WORKDIR /app

COPY . /app

RUN go build -o eastmoney main.go

EXPOSE 8000

CMD ["./eastmoney"]
