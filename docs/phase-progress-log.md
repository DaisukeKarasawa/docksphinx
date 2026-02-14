# Docksphinx MVP フェーズ進捗ログ

本ログは、フェーズ開始時/終了時の報告フォーマットを固定して記録する。

---

## フェーズ0: 現状把握と設計判断の固定

### 開始時（狙い/変更予定/リスク）
- 狙い: 実装前に仕様判断を固定し、後工程の手戻りを防ぐ。
- 変更予定: 現状コードの欠落分析、設定・TUI・stream・volume方針の確定。
- リスク: 入力指定ファイルの欠落、既存実装との不整合。

### 終了時
- What changed
  - 設計判断（TUI, 設定, daemon, backpressure, compose, volume代替）を固定。
- Where
  - `outputs/phase0-design-memo.md`（非追跡）
- How verified
  - `docs/requirements.md` と `internal/*` を照合
  - `go test ./...` 実行で現状破綻点を確認
- Remaining risks
  - `cmd/proto/api` 欠落により全体テスト不可

---

## フェーズ1: 縦切り（daemon→snapshot）

### 開始時（狙い/変更予定/リスク）
- 狙い: daemon起動とsnapshot取得の最短経路を成立させる提案作成。
- 変更予定: config/proto/daemon/CLI snapshotの差分提案。
- リスク: gRPC契約未確定のまま進めると再作業。

### 終了時
- What changed
  - phase1実装手順 + phase1-2 diff提案を作成。
- Where
  - `outputs/phase1-implementation.md`, `outputs/patches/phase1-2.diff.md`
- How verified
  - 依存関係とコンパイル阻害要因を静的確認
- Remaining risks
  - 実コード未適用（運用ルール制約）

---

## フェーズ2: 運用（launchd/log/history/tail）

### 開始時（狙い/変更予定/リスク）
- 狙い: 常駐運用と長時間安定性の提案を固める。
- 変更予定: launchd手順、tailのリーク防止方針。
- リスク: slow subscriber時の詰まり設計不足。

### 終了時
- What changed
  - 運用手順と backpressure 方針を提案に反映。
- Where
  - `outputs/phase2-implementation.md`, `outputs/ops/launchd.md`
- How verified
  - 設計レビュー（context/close/unsubscribe順序）
- Remaining risks
  - 実稼働検証は適用後に必要

---

## フェーズ3: TUI

### 開始時（狙い/変更予定/リスク）
- 狙い: 4ペイン+操作要件を満たす仕様固定。
- 変更予定: キーマップ、描画更新、pause仕様の定義。
- リスク: 描画更新と受信更新の競合。

### 終了時
- What changed
  - TUI操作仕様と実装差分提案を作成。
- Where
  - `outputs/phase3-implementation.md`, `outputs/specs/tui-interaction.md`, `outputs/patches/phase3-6.diff.md`
- How verified
  - 要件項目（Tab/矢印/jk/検索/ソート/pause/q）を照合
- Remaining risks
  - UI実動作の最終確認は適用後

---

## フェーズ4: 収集拡張

### 開始時（狙い/変更予定/リスク）
- 狙い: images/networks/volumes + uptimeの要件整合。
- 変更予定: メトリクス定義と欠損表示ルールの固定。
- リスク: volume usageの厳密取得不可能性。

### 終了時
- What changed
  - 収集拡張方針とOS差分を仕様化。
- Where
  - `outputs/phase4-implementation.md`, `outputs/specs/metrics-definition.md`
- How verified
  - Docker API制約と要件の突合
- Remaining risks
  - volumeは metadata-based 代替

---

## フェーズ5: 品質ゲート

### 開始時（狙い/変更予定/リスク）
- 狙い: テスト戦略とE2E方針を固定。
- 変更予定: test gate定義と実測ログ。
- リスク: proto欠落により全体テストが通らない。

### 終了時
- What changed
  - テスト計画と実行結果を記録。
- Where
  - `outputs/phase5-validation.md`, `outputs/test-plan.md`
- How verified
  - `go test` / `go test -race`（限定範囲）実行
- Remaining risks
  - 全体ゲートはphase1適用後に再実行が必要

---

## フェーズ6: Security / Refactoring

### 開始時（狙い/変更予定/リスク）
- 狙い: 脆弱性・静的解析・最小リファクタ方針を固定。
- 変更予定: 実行結果の証跡化と未解決リスク整理。
- リスク: ツール内エラーで全件完走できない可能性。

### 終了時
- What changed
  - `govulncheck/staticcheck/gosec` の結果を整理。
  - 修正方針（G115/deprecated等）を提案化。
- Where
  - `outputs/security-report.md`, `outputs/refactor-report.md`, `docs/security-check-summary.md`
- How verified
  - 各コマンドを実行し出力を記録
- Remaining risks
  - govulncheck/gosec の一部は環境要因で完走不可
