# Docksphinx MVP 制約事項と次の一手

## 目的
MVP時点での制約を明確化し、適用後に優先的に取り組む改善順序を示す。

---

## 1. MVP時点の制約

1. **直接コード変更禁止ルール**
   - 本作業では実装提案を `outputs/` で提供
   - 実コード適用は別工程

2. **volume usage の厳密取得**
   - Docker APIのみではポータブルな厳密使用量取得が難しい
   - MVPは metadata-based 代替

3. **compose依存可視化**
   - ラベル/ネットワークベースの推定であり、厳密解析ではない

4. **脆弱性ツールの完走性**
   - `govulncheck` / `gosec` は環境依存で失敗する可能性がある

---

## 2. 優先度付き 次の一手

## P0（最優先）
1. `outputs/patches/phase1-2.diff.md` を適用して proto/cmd/api を成立させる
2. `go test ./...` を通し、基礎動作（daemon/snapshot/tail）を実証する

## P1
1. `outputs/patches/phase3-6.diff.md` を適用し、TUIと収集拡張を実装
2. `go test -race ./...` でリーク/競合を検証

## P2
1. `govulncheck/staticcheck/gosec` をCIに組み込み再現性を担保
2. G115対応（型見直し/安全変換）を優先修正

## P3
1. volume usageの将来拡張（driver別戦略）
2. compose推定精度の計測と改善

---

## 3. 完了判定（適用後）
- `docksphinxd` 起動/停止/状態確認が動作
- `snapshot/tail/tui` が要件通り動作
- `go test ./...` + `go test -race ./...` が通る
- セキュリティチェック結果が運用可能な水準になる
