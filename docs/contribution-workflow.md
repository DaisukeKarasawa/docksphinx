# Docksphinx MVP Contribution Workflow

## 目的
MVP適用作業の標準フローを定義し、手順のばらつきを減らす。

---

## 1. 事前確認
1. `README.md` から関連docsへ到達
2. `docs/quickstart.md` と `docs/patch-application-procedure.md` を確認
3. `docs/risk-register.md` でHighリスクを確認

---

## 2. 実装適用フロー
1. phase1-2 パッチ適用
2. proto生成
3. 基本テスト実行
4. phase3-6 パッチ適用
5. race/セキュリティ/静的解析実行

---

## 3. 記録フロー
1. `docs/validation-log-template.md` でログ記録
2. 問題があれば `docs/risk-register.md` を更新
3. 必要なら `docs/limitations-and-next-steps.md` を更新

---

## 4. レビューフロー
1. `docs/reviewer-checklist.md` で要件適合チェック
2. `docs/requirements-traceability.md` で対応確認
3. `docs/plan-completion-checklist.md` でフェーズ証跡確認

---

## 5. 完了フロー
1. `docs/final-status.md` を更新
2. `docs/execution-changelog.md` に主要変更を追記
3. README導線が崩れていないか `docs/link-audit.md` で確認
