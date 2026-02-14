# Docksphinx MVP レビュアーチェックリスト

## 目的
MVP提案（requirements + docs + outputs）を短時間で検証するための確認項目。

---

## 1. スコープ適合
- [ ] `docs/requirements.md` がMVP範囲のみを扱っている
- [ ] 実行系操作（kill/restart等）がMVPに含まれていない
- [ ] 将来拡張（MCP/クラウド送信）がMVP実装対象に入っていない

---

## 2. 仕様固定
- [ ] `docs/mvp-spec-freeze.md` に設計判断が明記されている
- [ ] volume usage の代替指標方針が明記されている
- [ ] compose推定がヒューリスティックであることが明記されている

---

## 3. 受け入れ基準対応
- [ ] `docs/mvp-acceptance-matrix.md` で A〜G / H1〜H6 / Security / Refactor が対応付けられている
- [ ] 追跡先が `outputs` の成果物に結び付いている
- [ ] `docs/requirements-traceability.md` で要件対応が追跡できる

---

## 4. 適用手順の実行可能性
- [ ] `docs/patch-application-procedure.md` の順序で適用可能
- [ ] `docs/quickstart.md` で最短手順が明確
- [ ] `docs/implementation-handoff-runbook.md` の成功条件が明確
- [ ] ロールバック方針（フェーズ単位revert）が定義されている

---

## 5. 検証とセキュリティ
- [ ] `docs/security-check-summary.md` に実行コマンドと結果がある
- [ ] `govulncheck/staticcheck/gosec` の未解決点に回避策がある
- [ ] `go test` / `go test -race` の実測結果が記録されている

---

## 6. 認識論的明確性
- [ ] `docs/epistemic-assumptions.md` に Fact/Heuristic 分離がある
- [ ] 結論が覆る条件（Falsifiability）が記載されている

---

## 7. 最終導線
- [ ] `README.md` に主要docsへのリンクが揃っている
- [ ] レビュアーが README だけで必要文書へ到達できる
- [ ] `docs/review-sequence.md` の順序で確認できる
- [ ] `docs/plan-execution-report.md` と `docs/master-checklist.md` で最終判定できる
- [ ] `docs/acceptance-signoff-template.md` で判定結果を記録できる
