FROM golang:1-alpine AS build
WORKDIR /app

# add build dependencies for go-sqlite3
RUN apk add --no-cache \
	gcc \
	musl-dev

RUN go install github.com/air-verse/air@latest

COPY go.mod go.[sum] ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o main ./cmd/server

CMD ["air", "-c", ".air.toml"]
