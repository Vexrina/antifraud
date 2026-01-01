.PHONY: proto build run clean swagger

export PATH := $(LOCAL_BIN):$(PATH)

PROTO_ROOT=../third_party/googleapis
PROTO_INCLUDES=-I api -I $(PROTO_ROOT)
# Установка зависимостей
deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go get -u github.com/go-jet/jet/v2
	go install github.com/go-jet/jet/v2/cmd/jet@latest
	go install go.uber.org/mock/mockgen@latest
	export PATH=$PATH:$(go env GOPATH)/bin
	@echo "Dependencies installed"

# Генерация Go кода из proto файлов
proto: deps
	mkdir -p pkg/processing_core
	protoc \
	$(PROTO_INCLUDES) \
	--go_out=pkg --go-grpc_out=pkg \
	api/core.proto
	@echo "Proto files generated"



# Генерация swagger/openapi документации
swagger: proto
	protoc \
	$(PROTO_INCLUDES) \
	--openapiv2_out=pkg/core \
	--openapiv2_opt logtostderr=true \
	api/core.proto

# Сборка проекта
build: proto
	go build -o bin/server cmd/server/main.go
	@echo "Project built"

# Запуск сервера
run: build
	./bin/server

# Очистка
clean:
	rm -rf bin/
	rm -rf pkg/core/*.pb.go
	rm -rf pkg/core/*.swagger.json
	@echo "Clean complete"

# Генерация схем используя jet
# jet: deps
# 	jet -source=postgresql -host=localhost -port=5432 -user=processing_core_user -password=processing_core_password -dbname=processing_core_db -schema=public -path=generated/

# PROTO_DIR=../float-weaver/api
# PROTO_FILE=floatweaver.proto
# OUT_DIR=./pkg/floatweaver

# gen-floatweaver:
# 	mkdir -p $(OUT_DIR)
# 	protoc \
# 	  --proto_path=$(PROTO_DIR) \
# 	  --go_out=$(OUT_DIR) \
# 	  --go-grpc_out=$(OUT_DIR) \
# 	  $(PROTO_DIR)/$(PROTO_FILE)

# Полная генерация (все зависимости + очистка + генерация)
generate: clean deps proto swagger
	@echo "Full generation completed"
