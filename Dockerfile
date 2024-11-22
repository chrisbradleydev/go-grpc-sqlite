FROM golang:1-alpine AS build
WORKDIR /app

ARG BUILD_DIR

RUN if [ "${BUILD_DIR}" = "cmd/server" ]; then \
		apk add --no-cache gcc musl-dev; \
	fi

COPY go.mod go.[sum] ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=$([ "${BUILD_DIR}" = "cmd/server" ] && echo 1 || echo 0) \
    && GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main ./${BUILD_DIR}

FROM alpine:3 AS prod
WORKDIR /app

COPY --from=build /app/main /app/main

CMD ["/app/main"]
