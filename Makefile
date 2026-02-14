.PHONY: build clean test test-race proto generate install deps security quality help

# 変数定義
BINARY_DOCKSPHINX=./bin/docksphinx
BINARY_DOCKSPHINXD=./bin/docksphinxd
PROTO_DIR=./proto
API_DIR=./api
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*' -not -path './api/*')
# protoc と Go プラグインを探す PATH（GOPATH/bin, Homebrew 等）
export PATH := $(shell go env GOPATH)/bin:/opt/homebrew/bin:/usr/local/bin:$(PATH)
PROTOC ?= protoc

# デフォルトターゲット
.DEFAULT_GOAL := help

# ヘルプ表示
help:
	@echo "利用可能なコマンド"
	@echo " make build     - バイナリをビルド"
	@echo " make clean     - ビルド成果物を削除"
	@echo " make test      - テストを実行"
	@echo " make proto     - Protocol Buffers から Goコードを生成"
	@echo " make generate  - コード生成を実行"
	@echo " make install   - バイナリをインストール"
	@echo " make deps      - 依存関係を更新"
	@echo " make test-race - race detector 付きでテスト"
	@echo " make security  - staticcheck / gosec / govulncheck を実行"
	@echo " make quality   - test / race / security をまとめて実行"

# バイナリのビルド
build: proto
	@echo "Building binaries..."
	@mkdir -p ./bin
	go build -o $(BINARY_DOCKSPHINX) ./cmd/docksphinx
	go build -o $(BINARY_DOCKSPHINXD) ./cmd/docksphinxd
	@echo "Build complete: $(BINARY_DOCKSPHINX), $(BINARY_DOCKSPHINXD)"

# クリーンアップ
clean:
	@echo "Cleaning..."
	rm -rf ./bin
	rm -rf $(API_DIR)
	go clean
	@echo "Clean complete"

# テスト実行
test:
	@echo "Running tests..."
	go test -v ./...

test-race:
	@echo "Running race tests..."
	go test -race ./...

# Protocol BuffersからGoコードを生成
proto:
	@command -v $(PROTOC) >/dev/null 2>&1 || { echo "Error: protoc not found. Install with: brew install protobuf"; exit 127; }
	@echo "Generating gRPC code..."
	@mkdir -p $(API_DIR)/docksphinx/v1
	$(PROTOC) --go_out=$(API_DIR) \
	       --go_opt=paths=source_relative \
	       --go-grpc_out=$(API_DIR) \
	       --go-grpc_opt=paths=source_relative \
	       --proto_path=$(PROTO_DIR) \
	       $(PROTO_DIR)/docksphinx/v1/*.proto
	@echo "gRPC code generation complete"

# コード生成（proto + その他）
generate: proto
	@echo "Code generation complete"

# 依存関係の更新
deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
	@echo "Dependencies updated"

security: build
	@echo "Running staticcheck..."
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..."; go install honnef.co/go/tools/cmd/staticcheck@latest; }
	"$(shell go env GOPATH)/bin/staticcheck" ./...
	@echo "Running gosec..."
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest; }
	"$(shell go env GOPATH)/bin/gosec" -exclude-dir=api ./...
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	@echo "Running govulncheck in binary mode..."
	@"$(shell go env GOPATH)/bin/govulncheck" -mode=binary $(BINARY_DOCKSPHINX)
	@"$(shell go env GOPATH)/bin/govulncheck" -mode=binary $(BINARY_DOCKSPHINXD)
	@echo "Running govulncheck..."
	@out=$$(mktemp); \
	if "$(shell go env GOPATH)/bin/govulncheck" ./... >$$out 2>&1; then \
		cat $$out; \
	else \
		cat $$out; \
		if rg -q "internal error:" $$out; then \
			echo "WARNING: govulncheck internal error detected. See docs/security-check-summary.md"; \
		else \
			echo "ERROR: govulncheck failed with non-internal error"; \
			rm -f $$out; \
			exit 1; \
		fi; \
	fi; \
	rm -f $$out

quality: test test-race security
	@echo "Quality gates complete"

# インストール（GOPATH/binにインストール）
install: build
	@echo "Installing binaries..."
	go install ./cmd/docksphinx
	go install ./cmd/docksphinxd
	@echo "Installation complete"

# 開発用: ホットリロード（air等を使用する場合）
dev:
	@echo "Starting development mode..."
	@echo "No built-in hot reload target. Use your preferred tool (e.g. air) with docksphinx commands."
