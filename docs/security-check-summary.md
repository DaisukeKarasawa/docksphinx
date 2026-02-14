# Security / Vulnerability チェック結果サマリ（実行証跡）

## 実行日時
- 2026-02-14

---

## 1. 実行コマンド

```bash
go test ./...
go test ./internal/docker ./internal/event ./internal/monitor
go test -race ./internal/docker ./internal/event ./internal/monitor
govulncheck ./...
staticcheck ./...
gosec ./...
```

---

## 2. 結果要約

## `go test ./...`
- 失敗
- 主因: `docksphinx/api/docksphinx/v1` が存在しない（proto生成物不足）

## `go test`（限定パッケージ）
- 成功
- 対象: `internal/docker`, `internal/event`, `internal/monitor`

## `go test -race`（限定パッケージ）
- 成功

## `govulncheck ./...`
- 失敗
- 事象:
  - ツール未導入 → 導入後再試行
  - 再試行時、依存解析内部エラー（docker依存の型情報解決に失敗）

## `staticcheck ./...`
- 失敗（検出あり）
- 主な指摘:
  - ST1005: エラー文字列の先頭大文字
  - SA1019: deprecated API使用
  - SA4006: 未使用代入
  - grpc生成物不足によるコンパイルエラー

## `gosec ./...`
- 失敗（検出あり）
- 主な指摘:
  - G115: `uint64 -> int64` 変換で overflow懸念（metrics周辺）
  - grpcパッケージ解析時のSSA panic

---

## 3. リスク評価
- 現段階の最大リスクは「実装未適用（cmd/proto/api欠落）」による全体ゲート未達。
- セキュリティ観点では、数値変換（G115）と静的解析指摘を phase6パッチ適用で先に解消する必要がある。

---

## 4. 次アクション
1. `outputs/patches/phase1-2.diff.md` を適用
2. `make proto` 実行
3. `go test ./...` 再実行
4. `outputs/patches/phase3-6.diff.md` 適用
5. `go test -race ./...`, `govulncheck`, `staticcheck`, `gosec` 再実行
