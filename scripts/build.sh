#!/bin/bash
set -e

# カラー出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Docksphinx...${NC}"

# プロジェクトルートに移動
cd "$(dirname "$0")/.."

# バイナリディレクトリの作成
mkdir -p bin

# Protocol Buffers の生成
if [ -d "proto" ] && [ "$(ls -A proto)" ]; then
  echo -e "${YELLOW}Generating gRPC code...${NC}"
  make proto || echo -e "${YELLOW}Warning: proto generation skipped${NC}"
fi

# ビルド
echo -e "${GREEN}Building docksphinx...${NC}"
go build -o bin/docksphinx ./cmd/docksphinx

echo -e "${GREEN}Building docksphinxd...${NC}"
go build -o bin/docksphinxd ./cmd/docksphinxd

echo -e "${GREEN}Build complete!${NC}"
echo -e "Binaries are in: ${GREEN}./bin/${NC}"
