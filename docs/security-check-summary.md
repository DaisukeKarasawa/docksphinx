# Security Check Summary (MVP)

最終実装時点のセキュリティ/静的解析結果を記録します。

## 実行コマンド

```bash
go test ./...
go test -race ./...
"$(go env GOPATH)/bin/staticcheck" ./...
"$(go env GOPATH)/bin/gosec" -exclude-dir=api ./...
"$(go env GOPATH)/bin/govulncheck" -mode=binary ./bin/docksphinx
"$(go env GOPATH)/bin/govulncheck" -mode=binary ./bin/docksphinxd
"$(go env GOPATH)/bin/govulncheck" ./...
```

## 結果

- `go test ./...` : PASS
- `go test -race ./...` : PASS
- `staticcheck ./...` : PASS
- `gosec -exclude-dir=api ./...` : PASS（Issues: 0）
- `govulncheck -mode=binary` : PASS（`docksphinx`, `docksphinxd` ともに No vulnerabilities found）
- `govulncheck ./...` : FAIL（ツール内部エラー）

### govulncheck 失敗内容

```
internal error: package "golang.org/x/text/encoding" without types was imported from "github.com/gdamore/encoding"
```

## 実施した切り分け

1. `go mod tidy` 後に再実行
2. `go list -deps ./...` で依存を先解決して再実行
3. `govulncheck` バージョン固定（`v1.1.4`）で再実行
4. `GOTOOLCHAIN=go1.24.11` 固定で再実行
5. `govulncheck -mode=binary` で実行（内部エラー回避）

1〜4 は同一エラー再現。5 は成功。

## 判断

- source mode 失敗はアプリ実装ではなく `govulncheck` 実行環境/依存解決相性に起因する可能性が高い。
- 代替として binary mode の `govulncheck`、`staticcheck`、`gosec`、`go test -race` を通過済み。

## 修正済み脆弱性対応

- binary mode で検出された Go 標準ライブラリ脆弱性（GO-2026-4341 / GO-2026-4340 / GO-2026-4337）に対して、
  `toolchain go1.24.11` から `toolchain go1.24.13` へ更新し、再スキャンで解消を確認。

## 次アクション候補

- CI で別 Go バージョン（例: 1.24.x / 1.25.x）マトリクスで `govulncheck` 再試行
- `tview`/`tcell` 依存を含む最小再現で upstream issue 化
