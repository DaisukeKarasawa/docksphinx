# Docksphinx MVP リスク台帳

## 目的
MVP適用前後のリスクをID管理し、優先度と対策方針を明確化する。

---

| ID | リスク | 影響 | 優先度 | 現在の対策 | 次アクション |
|---|---|---|---|---|---|
| R-001 | 実コード未適用（運用ルール制約） | 実行確認が別工程化 | High | outputsに具体diffを整備 | 実装担当がphase1-2→phase3-6を適用 |
| R-002 | `go test ./...` がproto欠落で失敗 | 全体品質ゲート未達 | High | phase1-2でproto/cmd/api整備提案 | `make proto` 後に全体再テスト |
| R-003 | `govulncheck` 内部エラー | 脆弱性網羅確認が未完 | Medium | 実行証跡と回避策を文書化 | toolchain固定して再実行 |
| R-004 | `gosec` G115（型変換） | 極値でoverflow懸念 | Medium | 修正方針を提案 | uint64統一または安全変換導入 |
| R-005 | volume usageが代替指標 | 厳密容量は把握不可 | Medium | requirementsに制約明記 | 将来拡張でdriver別戦略 |
| R-006 | compose推定の誤判定 | 依存可視化の精度低下 | Low | ラベル優先+制約明記 | 運用データで精度評価 |

---

## 更新ルール
- 重要な検証結果が出たら本台帳を更新する。
- Highの項目は、次の実装/検証サイクルで必ず進捗を作る。
