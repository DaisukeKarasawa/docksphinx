#!/bin/bash
set -e

# カラー出力
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Installing Docksphinx...${NC}"

# プロジェクトルートに移動
cd "$(dirname "$0")/.."

# 依存関係のインストール
echo -e "${GREEN}Installing dependencies...${NC}"
go mod download
go mod tidy

# Protocol Buffersツールのインストール
echo -e "${GREEN}Installing Protocol Buffers tools...${NC}"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# ビルド
echo -e "${GREEN}Building...${NC}"
make build

# インストール
echo -e "${GREEN}Installing binaries to GOPATH/bin...${NC}"
make install

echo -e "${GREEN}Installation complete!${NC}"
echo -e "You can now use ${GREEN}docksphinx${NC} and ${GREEN}docksphinxd${NC} commands"
