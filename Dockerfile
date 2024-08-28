FROM golang:1.23-alpine AS build

WORKDIR /app

RUN apk update && apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN go build -o goeff-bot .

FROM scratch

WORKDIR /app

COPY --from=build /app/goeff-bot .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/goeff-bot"]
