include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

# confirm changes
.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #
.DEFAULT_GOAL := run

# run the cmd/web application
.PHONY: run
run:
	APP_ENV=${APP_ENV} \
	GRPC_HOST=${GRPC_HOST} \
	GRPC_PORT=${GRPC_PORT} \
	SQLITE_DB=${SQLITE_DB} air

# generate protos
protos:
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		${PROTO_DIR}/*.proto
