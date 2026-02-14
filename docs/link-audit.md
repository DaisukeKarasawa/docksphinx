# ドキュメントリンク監査（README）

## 実施日
- 2026-02-14

## 目的
READMEに記載された `docs/*` 参照リンクが、実在ファイルと一致していることを確認する。

---

## 1. 監査対象
README の「主要ドキュメント」節にある以下リンク:

- `docs/requirements.md`
- `docs/mvp-spec-freeze.md`
- `docs/mvp-acceptance-matrix.md`
- `docs/mvp-execution-report.md`
- `docs/implementation-handoff-runbook.md`
- `docs/security-check-summary.md`
- `docs/phase-progress-log.md`
- `docs/epistemic-assumptions.md`
- `docs/patch-application-procedure.md`
- `docs/reviewer-checklist.md`
- `docs/command-reference.md`
- `docs/delivery-manifest.md`
- `docs/adr-0001-mvp-delivery-mode.md`

---

## 2. 監査結果
- 判定: **PASS**
- 理由:
  - READMEから抽出した全リンクについて、`docs/` 内に同名ファイルの存在を確認。

---

## 3. 監査手順（再現用）

```bash
rg "`docs/[^`]+`" README.md
ls -1 docs
```

---

## 4. 次回監査トリガー
- READMEの主要ドキュメント一覧を更新したとき
- docs配下ファイル名を変更したとき
