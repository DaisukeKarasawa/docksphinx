# Docksphinx MVP 成果物マニフェスト

## 目的
本タスクで作成した成果物の所在と役割を一目で追跡できるようにする。

---

## 1. Git追跡対象（docs）

- `docs/requirements.md`
  - MVP要件本体（固定仕様追記済み）
- `docs/mvp-spec-freeze.md`
  - 設計判断の固定
- `docs/mvp-execution-report.md`
  - 実行結果（運用ルール準拠）
- `docs/mvp-acceptance-matrix.md`
  - A〜G/H1〜H6/Security/Refactor の対応表
- `docs/implementation-handoff-runbook.md`
  - 実装担当への引き継ぎ手順
- `docs/security-check-summary.md`
  - セキュリティ/検証コマンド結果サマリ
- `docs/phase-progress-log.md`
  - フェーズ開始/終了ログ
- `docs/epistemic-assumptions.md`
  - 事実・経験則・反証条件
- `docs/patch-application-procedure.md`
  - outputsパッチ適用実行手順
- `docs/reviewer-checklist.md`
  - レビュー観点チェックリスト
- `docs/command-reference.md`
  - 運用コマンド集
- `docs/index.md`
  - ドキュメント索引
- `docs/quickstart.md`
  - 最短適用手順
- `docs/final-status.md`
  - 現時点の完了/未完要約
- `docs/risk-register.md`
  - リスク台帳
- `docs/validation-log-template.md`
  - 検証ログ記録テンプレ
- `docs/plan-completion-checklist.md`
  - PLAN完了証跡
- `docs/requirements-traceability.md`
  - 要件と成果物の対応表
- `docs/contribution-workflow.md`
  - 参加者向け標準フロー
- `docs/incident-playbook.md`
  - 障害初動手順
- `docs/macos-launchd-checklist.md`
  - launchd運用確認項目
- `docs/security-baseline-checklist.md`
  - セキュリティ基準確認項目
- `docs/minimal-regression-suite.md`
  - 最小回帰テスト
- `docs/acceptance-signoff-template.md`
  - 受け入れサインオフ雛形
- `docs/review-sequence.md`
  - レビュー順序
- `docs/navigation-map.md`
  - 文書導線図
- `docs/maintenance-routine.md`
  - 定期運用項目
- `docs/owner-rotation-note.md`
  - 担当交代時メモ
- `docs/master-checklist.md`
  - 横断マスターチェック

---

## 2. Git非追跡成果（outputs）

- `outputs/missing-implementations.md`
- `outputs/phase0-design-memo.md`
- `outputs/phase1-implementation.md`
- `outputs/phase2-implementation.md`
- `outputs/phase3-implementation.md`
- `outputs/phase4-implementation.md`
- `outputs/phase5-validation.md`
- `outputs/phase6-security-refactor.md`
- `outputs/patches/phase1-2.diff.md`
- `outputs/patches/phase3-6.diff.md`
- `outputs/specs/metrics-definition.md`
- `outputs/specs/tui-interaction.md`
- `outputs/ops/launchd.md`
- `outputs/test-plan.md`
- `outputs/security-report.md`
- `outputs/refactor-report.md`
- `outputs/README-update-proposal.md`

---

## 3. 成果物整合性チェック
- README から主要docsへ遷移可能
- docs から outputs成果（適用対象）への導線が明示されている
- 要件・受け入れ基準・検証結果・未解決リスクが分離記述されている
