# Docksphinx MVP 実装適用ハンドオフランブック

## 目的
本リポジトリの運用ルール（直接コード変更禁止）に従って作成された `outputs/` 成果物を、実装担当者が迷わず適用できるようにする。

---

## 1. 前提（Premises）

## 検証可能な前提
1. 現行リポジトリには `cmd/`, `proto/`, `api/` が未作成である。
2. `go test ./...` は gRPC生成物不足で失敗する。
3. `outputs/` に段階的実装提案（phase0〜phase6 + patch案）が存在する。

## ヒューリスティック前提
1. TUIは Bubble Tea 採用が最小リスク。
2. volume usage はMVPで metadata-based代替が妥当。

---

## 2. 適用順序（厳守）

1. `outputs/patches/phase1-2.diff.md`
2. `make proto`
3. `go test ./...`
4. `outputs/patches/phase3-6.diff.md`
5. `go test ./...`
6. `go test -race ./...`
7. `govulncheck ./...`
8. `staticcheck ./...`
9. `gosec ./...`

---

## 3. 成功判定（Success State）
- `docksphinxd` 起動/停止/状態確認が可能
- `docksphinx snapshot/tail/tui` が機能
- TUI操作（Tab/矢印/jk/検索/ソート/pause/q）が機能
- 閾値イベントが連続N回 + cooldownで動作
- 収集対象（containers/images/networks/volumes）が表示される
- `go test ./...` が通る

---

## 4. 結論が覆る条件（Falsifiability）
以下が発生した場合、現在の実装提案は再設計が必要:

1. `phase1-2` 適用後も `GetSnapshot/Stream` 契約がコンパイル不能
2. `go test ./...` で、proto不足以外の構造的失敗が多数発生
3. TUI event loop が実運用でフリーズし、backpressure設計で回避不能
4. volume metadata代替がMVP運用要件を満たさないと判断された場合

---

## 5. 失敗時ロールバック方針
- フェーズ単位でコミットを分け、`phase1-2` と `phase3-6` を独立して戻せるようにする。
- 失敗したフェーズのコミットのみrevertし、前段階の動作状態へ戻す。

---

## 6. 参照ドキュメント
- `docs/requirements.md`
- `docs/mvp-spec-freeze.md`
- `docs/mvp-acceptance-matrix.md`
- `docs/mvp-execution-report.md`
