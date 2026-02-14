# CI Gate Template（MVP適用後向け）

## 目的
MVP実装差分適用後に、品質ゲートを自動実行するためのテンプレートを示す。

---

## 推奨ジョブ構成

1. `test`
   - `go test ./...`
2. `race`（必要に応じて並列）
   - `go test -race ./...`
3. `vuln`
   - `govulncheck ./...`
4. `static`
   - `staticcheck ./...`
5. `sec`
   - `gosec ./...`

---

## GitHub Actions 例

```yaml
name: mvp-quality-gate

on:
  push:
    branches: [ "cursor/**" ]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24.x"
      - run: go test ./...

  race:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24.x"
      - run: go test -race ./...

  vuln:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24.x"
      - run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - run: govulncheck ./...

  static:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24.x"
      - run: go install honnef.co/go/tools/cmd/staticcheck@latest
      - run: staticcheck ./...

  sec:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24.x"
      - run: go install github.com/securego/gosec/v2/cmd/gosec@latest
      - run: gosec ./...
```

---

## 注意点
- proto未生成の場合に `go test ./...` が落ちるため、必要なら先に `make proto` を追加する。
- `govulncheck` / `gosec` は依存やGoバージョンの組み合わせで失敗することがあるため、失敗ログを保存して再現性を確保する。
