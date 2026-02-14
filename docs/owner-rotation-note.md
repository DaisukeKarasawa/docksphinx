# オーナー交代時メモ（MVP運用）

## 目的
担当者が交代しても、MVP運用・適用・検証が止まらないよう最小限の引継ぎ項目を定義する。

---

## 1. まず共有する文書
- `README.md`
- `docs/final-status.md`
- `docs/risk-register.md`
- `docs/quickstart.md`
- `docs/patch-application-procedure.md`

---

## 2. 必須確認
- [ ] 現在の未解決リスク（High）を把握
- [ ] 直近の検証結果（security-check-summary, command-output-samples）を確認
- [ ] 次の優先アクション（limitations-and-next-steps）を確認

---

## 3. 引継ぎ会話で確認すべき質問
1. 現在の最大リスクは何か？
2. どの検証が未完か？
3. どの文書を更新すると全体整合が崩れないか？

---

## 4. 更新ルール
- READMEのリンクを更新したら、`docs/index.md` と `docs/link-audit.md` も更新する。
- 重要な方針変更があれば `AGENTS.md` と `final-status.md` を更新する。
