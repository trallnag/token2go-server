FROM golang:1.19-alpine AS build

ENV CGO_ENABLED 0

WORKDIR /build

COPY go.mod .
# COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o token2go-server

FROM alpine:3.16

COPY --from=build /build/token2go-server /usr/local/bin/

CMD [ "token2go-server" ]
