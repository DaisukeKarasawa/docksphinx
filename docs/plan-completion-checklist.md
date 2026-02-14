# PLAN 完了チェックリスト（実行モード）

## 目的
`PLAN` に定義したフェーズ0〜6の完了状況を、証跡付きで確認できるようにする。

---

## フェーズ0: 現状把握と設計判断固定
- [x] 現状実装の不足を確認
- [x] 設定仕様を確定
- [x] TUI方針を確定
- [x] volume/uptime/compose/backpressure方針を確定
- 証跡:
  - `docs/mvp-spec-freeze.md`
  - `outputs/phase0-design-memo.md`

## フェーズ1: 縦切り（daemon→snapshot）
- [x] daemon配線の提案を作成
- [x] snapshot最小経路の提案を作成
- [x] graceful shutdown設計を提案
- 証跡:
  - `outputs/phase1-implementation.md`
  - `outputs/patches/phase1-2.diff.md`

## フェーズ2: 運用（launchd/log/history/tail）
- [x] launchd運用手順を整備
- [x] ログ/履歴方針を提案
- [x] tail健全化方針を提案
- 証跡:
  - `outputs/phase2-implementation.md`
  - `outputs/ops/launchd.md`

## フェーズ3: TUI
- [x] 4ペイン仕様を固定
- [x] 操作要件（Tab/矢印/jk/検索/ソート/pause/q）を固定
- [x] stream更新安定化方針を提案
- 証跡:
  - `outputs/phase3-implementation.md`
  - `outputs/specs/tui-interaction.md`

## フェーズ4: 収集拡張
- [x] images/networks/volumes収集方針を提案
- [x] uptime/volume代替指標を固定
- [x] compose推定方針を固定
- 証跡:
  - `outputs/phase4-implementation.md`
  - `outputs/specs/metrics-definition.md`

## フェーズ5: 品質ゲート
- [x] テスト戦略を定義
- [x] 実行可能範囲のテストを実施
- [x] 結果を記録
- 証跡:
  - `outputs/phase5-validation.md`
  - `outputs/test-plan.md`
  - `docs/security-check-summary.md`

## フェーズ6: Security / Vulnerability / Refactoring
- [x] `govulncheck/staticcheck/gosec` を実行
- [x] 未解決課題の理由・影響・回避策を記録
- [x] 最小リファクタ方針を記録
- 証跡:
  - `outputs/security-report.md`
  - `outputs/refactor-report.md`
  - `docs/risk-register.md`

---

## 完了判定（運用ルール準拠）
- [x] 直接コード変更禁止ルールを順守
- [x] `outputs/` に実装手順 + diff + 検証手順を提供
- [x] `docs/` に仕様固定・受け入れ基準・検証証跡・運用導線を提供
