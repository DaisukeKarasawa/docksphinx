#!/bin/bash
set -e

# カラー出力
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Generating Protocol Buffers code...${NC}"

# プロジェクトルートに移動
cd "$(dirname "$0")/.."

# ディレクトリの定義
PROTO_DIR=./proto
API_DIR=./api

# APIディレクトリの作成
mkdir -p ${API_DIR}/docksphinx/v1

# Protocol Buffersコンパイラの確認
if ! command -v protoc &> /dev/null; then
  echo -e "${YELLOW}Error: protoc is not installed${NC}"
  echo "Please install Protocol Buffers compiler:"
  echo "  macOS: brew install protobuf"
  echo "  Linux: apt-get install protobuf-compiler"
  exit 1
fi

# Goプラグインの確認
if ! command -v protoc-gen-go &> /dev/null; then
  echo -e "${YELLOW}Installing protoc-gen-go...${NC}"
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
  echo -e "${YELLOW}Installing protoc-gen-go-grpc...${NC}"
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Protocol Buffersのコンパイル
if [ -f "${PROTO_DIR}/docksphinx/v1/docksphinx.proto" ]; then
  echo -e "${GREEN}Compiling ${PROTO_DIR}/docksphinx/v1/docksphinx.proto...${NC}"
  protoc --go_out=${API_DIR} \
         --go_opt=paths=source_relative \
         --go-grpc_out=${API_DIR} \
         --go-grpc_opt=paths=source_relative \
         --proto_path=${PROTO_DIR} \
         ${PROTO_DIR}/docksphinx/v1/docksphinx.proto
  echo -e "${GREEN}Code generation complete!${NC}"
else
  echo -e "${YELLOW}Warning: proto file not found. Skipping...${NC}"
fi