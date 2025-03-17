# syntax=docker/dockerfile:1

FROM golang:alpine AS builder

RUN mkdir -p /go/src/familycoin

WORKDIR /go/src/familycoin
COPY . .

ENV CGO_ENABLED=0
RUN go get \
  && go mod download \ 
  && go build -a -o ./familycoin

FROM alpine

WORKDIR /app

COPY --from=builder /go/src/familycoin /app

ENV GO111MODULE="on"

EXPOSE 8881
EXPOSE 8882

CMD ["./familycoin"]