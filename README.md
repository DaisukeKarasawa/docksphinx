# Docksphinx

Docker環境を読み取り専用で監視し、CLI/TUIで状態を可視化するツールのMVP仕様リポジトリです。

## 主要ドキュメント

- 要件定義: `docs/requirements.md`
- MVP仕様固定: `docs/mvp-spec-freeze.md`
- MVP受け入れ基準マトリクス: `docs/mvp-acceptance-matrix.md`
- 実行レポート（運用ルール準拠版）: `docs/mvp-execution-report.md`
- 実装適用ハンドオフランブック: `docs/implementation-handoff-runbook.md`
- セキュリティ/検証結果サマリ: `docs/security-check-summary.md`
- フェーズ進捗ログ: `docs/phase-progress-log.md`
- 前提と反証条件（Epistemic/Falsifiability）: `docs/epistemic-assumptions.md`
- パッチ適用手順（実行マニュアル）: `docs/patch-application-procedure.md`
- レビュアーチェックリスト: `docs/reviewer-checklist.md`
- コマンドリファレンス: `docs/command-reference.md`
- 成果物マニフェスト: `docs/delivery-manifest.md`
- ADR-0001（outputsベース実装提供方針）: `docs/adr-0001-mvp-delivery-mode.md`

## 補足

- このリポジトリでは運用ルールにより、実装提案は `outputs/` に段階的ドキュメントとして出力されます（`outputs/` は git 管理対象外）。
