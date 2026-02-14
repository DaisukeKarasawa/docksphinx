# 要件トレーサビリティ表（MVP）

## 目的
`docs/requirements.md` の主要要件が、どの成果物（docs/outputs）でカバーされているかを追跡する。

---

| 要件ID | 要件概要 | 対応成果（docs） | 対応成果（outputs） |
|---|---|---|---|
| REQ-A | daemon起動/停止/状態確認 | `mvp-acceptance-matrix.md`, `implementation-handoff-runbook.md` | `phase1-implementation.md`, `patches/phase1-2.diff.md` |
| REQ-B | 設定ロード/永続化 | `requirements.md`, `mvp-spec-freeze.md` | `phase1-implementation.md`, `patches/phase1-2.diff.md` |
| REQ-C | ログ/履歴 | `security-check-summary.md`, `risk-register.md` | `phase2-implementation.md`, `patches/phase1-2.diff.md` |
| REQ-D | snapshot/tail健全化 | `command-reference.md`, `handover-checklist.md` | `phase2-implementation.md`, `phase5-validation.md` |
| REQ-E | TUI 4ペイン+操作 | `requirements.md`, `mvp-acceptance-matrix.md` | `phase3-implementation.md`, `specs/tui-interaction.md` |
| REQ-F | image/network/volume収集 | `requirements.md`, `limitations-and-next-steps.md` | `phase4-implementation.md`, `specs/metrics-definition.md` |
| REQ-G | uptime / volume方針 | `requirements.md`, `mvp-spec-freeze.md` | `phase4-implementation.md`, `specs/metrics-definition.md` |
| REQ-H1 | launchd運用 | `command-reference.md`, `quickstart.md` | `ops/launchd.md`, `phase2-implementation.md` |
| REQ-H2 | backpressure/leak | `epistemic-assumptions.md`, `risk-register.md` | `phase2-implementation.md`, `patches/phase3-6.diff.md` |
| REQ-H3 | cooldown | `requirements.md` | `patches/phase3-6.diff.md` |
| REQ-H4 | compose推定 | `requirements.md`, `mvp-spec-freeze.md` | `phase4-implementation.md`, `patches/phase3-6.diff.md` |
| REQ-H5 | メトリクス定義 | `requirements.md`, `glossary.md` | `specs/metrics-definition.md` |
| REQ-H6 | テスト戦略 | `security-check-summary.md`, `ci-gate-template.md` | `phase5-validation.md`, `test-plan.md` |
| REQ-OPS1 | 障害初動/運用復旧 | `incident-playbook.md`, `macos-launchd-checklist.md` | `ops/launchd.md` |
| REQ-OPS2 | セキュリティ基準運用 | `security-baseline-checklist.md`, `risk-register.md` | `security-report.md` |
| REQ-OPS3 | 回帰運用固定 | `minimal-regression-suite.md`, `validation-log-template.md` | `phase5-validation.md` |
| REQ-OPS4 | 承認・サインオフ | `review-sequence.md`, `acceptance-signoff-template.md`, `master-checklist.md` | `phase6-security-refactor.md` |
| REQ-OPS5 | 継続保守 | `maintenance-routine.md`, `owner-rotation-note.md`, `navigation-map.md` | `README-update-proposal.md` |

---

## 補足
- 本表は「適用前提の成果物対応」を示す。  
- 実コードの最終適用は `patch-application-procedure.md` に従って別工程で行う。
