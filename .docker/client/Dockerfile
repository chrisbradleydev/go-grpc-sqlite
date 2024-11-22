FROM golang:1-alpine AS build
WORKDIR /app

COPY go.mod go.[sum] ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main ./cmd/client

FROM alpine:3 AS prod
WORKDIR /app

COPY --from=build /app/main /app/main

CMD ["/app/main"]
