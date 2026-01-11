.PHONY: proto build run clean swagger

export PATH := $(LOCAL_BIN):$(PATH)

PROTO_INCLUDES=-I api
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

proto_kafka:
	mkdir -p pkg/processing_core
	protoc \
		$(PROTO_INCLUDES) \
		--go_out=pkg \
		api/kafka.proto
	@echo "Kafka Proto files generated"

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
jet: deps
	jet -source=postgresql -host=localhost -port=5433 -user=proc_core_user -password=proc_core_pwd -dbname=proc_core_db -schema=public -path=generated/

PROTO_DIR=../antifraud/api
PROTO_FILE=antifraud.proto
OUT_DIR=./pkg

gen-antifraud:
	mkdir -p $(OUT_DIR)
	protoc \
		-I $(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		$(PROTO_FILE)

# Полная генерация (все зависимости + очистка + генерация)
generate: clean deps proto swagger proto_kafka antifraud jet
	@echo "Full generation completed"
