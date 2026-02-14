# Docksphinx MVP 障害初動プレイブック

## 目的
長時間稼働時に問題が起きた際、最短で「状況把握 → 原因切り分け → 暫定対処」に進むための手順。

---

## 0. 初動3分ルール
1. 現象を1行で定義（例:「tail更新が止まった」）
2. 再現条件を固定（いつ・どのコマンドで・何分後）
3. 影響範囲を確認（daemonのみ / CLIのみ / 全体）

---

## 1. 症状別チェック

## A) snapshot は動くが tail/tui が止まる
- 確認:
  - gRPC streamの接続維持状態
  - subscriber詰まり（backpressure）
- コマンド:
  ```bash
  ./bin/docksphinx snapshot --addr 127.0.0.1:50051
  ./bin/docksphinx tail --addr 127.0.0.1:50051
  ```
- 想定原因:
  - stream再接続ロジック不足
  - broadcaster drop過多

## B) daemon がすぐ終了する
- 確認:
  - PIDファイルの整合
  - 設定ファイル値（address/interval/regex）
- コマンド:
  ```bash
  ./bin/docksphinx status
  ```
- 想定原因:
  - 設定バリデーション失敗
  - Docker接続失敗

## C) CPU/Mem 閾値イベントが過剰発火する
- 確認:
  - consecutive_count
  - cooldown設定
- 想定原因:
  - ノイズ対策不足
  - 閾値設定が低すぎる

## D) volume が N/A で困る
- 確認:
  - MVP仕様（metadata-only）
- 対応:
  - 要件外であることを再確認し、将来拡張タスク化

---

## 2. 収集すべき証跡
- 実行コマンドと時刻
- 失敗ログ（標準出力/標準エラー）
- `go test` / `race` / `staticcheck` / `gosec` の結果
- 再現手順（最短）

記録先:
- `docs/validation-log-template.md`
- `docs/security-check-summary.md`
- `docs/risk-register.md`

---

## 3. 暫定対処の原則
- まず観測性を上げる（ログ/状態確認）
- 次に被害局所化（tail再接続、不要購読停止）
- 恒久対処は要件スコープ内で最小修正

---

## 4. 収束判定
- 再現手順で再発しない
- 監視操作（snapshot/tail/tui）が継続動作する
- 追加リスクを `risk-register` に反映済み
