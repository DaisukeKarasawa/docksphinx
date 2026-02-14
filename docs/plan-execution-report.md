# PLAN 実行レポート（Evidence Index）

## 目的
PLANの実行状況を、フェーズ単位で「状態・根拠・残課題」に分けて確認できるようにする。

---

## 1. フェーズ進捗

| フェーズ | 状態 | 根拠 |
|---|---|---|
| 0: 現状把握/設計固定 | 完了 | `docs/mvp-spec-freeze.md`, `outputs/phase0-design-memo.md` |
| 1: 縦切り（daemon→snapshot） | 提案完了 | `outputs/phase1-implementation.md`, `outputs/patches/phase1-2.diff.md` |
| 2: 運用整備（launchd/log/history/tail） | 提案完了 | `outputs/phase2-implementation.md`, `outputs/ops/launchd.md` |
| 3: TUI | 提案完了 | `outputs/phase3-implementation.md`, `outputs/specs/tui-interaction.md` |
| 4: 収集拡張 | 提案完了 | `outputs/phase4-implementation.md`, `outputs/specs/metrics-definition.md` |
| 5: 品質ゲート | 実行記録あり | `outputs/phase5-validation.md`, `docs/security-check-summary.md` |
| 6: Security/Refactor | 実行記録あり | `outputs/security-report.md`, `outputs/refactor-report.md` |
| 7: 運用ドキュメント強化 | 完了 | `docs/phase-progress-log.md`（フェーズ7節） |

---

## 2. Hard Tasks 状態

| Task | 状態 | 根拠 |
|---|---|---|
| H1 launchd | 提案完了 | `outputs/ops/launchd.md`, `docs/macos-launchd-checklist.md` |
| H2 leak/backpressure | 提案完了 | `outputs/patches/phase1-2.diff.md`, `outputs/patches/phase3-6.diff.md` |
| H3 cooldown | 提案完了 | `outputs/patches/phase3-6.diff.md`, `docs/requirements.md` |
| H4 compose grouping | 提案完了 | `outputs/phase4-implementation.md`, `docs/requirements.md` |
| H5 metrics definition | 固定済み | `outputs/specs/metrics-definition.md`, `docs/mvp-spec-freeze.md` |
| H6 test strategy | 定義＋部分実施 | `outputs/test-plan.md`, `docs/security-check-summary.md` |

---

## 3. Security / Refactor Gate 状態

### Security
- 実施:
  - `govulncheck`, `staticcheck`, `gosec`（実行証跡あり）
- 根拠:
  - `docs/security-check-summary.md`
  - `docs/command-output-samples.md`
- 残課題:
  - toolchain/依存要因で完走不可のケースあり（`risk-register.md`）

### Refactor
- 実施:
  - 最小リファクタ方針を文書化
- 根拠:
  - `outputs/refactor-report.md`
  - `docs/requirements-traceability.md`

---

## 4. 未解決事項（運用ルール起因）

1. 実コード未適用（直接コード変更禁止）
2. `go test ./...` 全体は proto/cmd/api 適用後に再判定
3. 一部セキュリティツールの完走性に環境依存あり

---

## 5. 次アクション（適用担当向け）
1. `outputs/patches/phase1-2.diff.md` 適用
2. `make proto` → `go test ./...`
3. `outputs/patches/phase3-6.diff.md` 適用
4. `go test -race ./...`
5. `govulncheck/staticcheck/gosec` 再実行
6. `docs/acceptance-signoff-template.md` で判定記録
