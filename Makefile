PROTO_FILES := proxy/proxy.proto
PROTOC ?= protoc

.PHONY: build test vet proto cast

build:
	go build -o emqutiti ./cmd/emqutiti

vet:
	go vet ./...

test: vet
	go test ./...

proto:
	$(PROTOC) --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. $(PROTO_FILES)

cast:
	docker build -f docs/scripts/Dockerfile.cast -t emqutiti-cast .
	docker run --rm -it \
		-v "$(PWD)/docs:/app/docs" \
		emqutiti-cast docs/scripts/record_casts.sh
