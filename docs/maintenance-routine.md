# Docksphinx MVP 定期メンテナンス運用

## 目的
長時間運用での劣化を早期に発見するため、定期確認項目を固定する。

---

## 週次ルーチン
- [ ] `go test ./...`（適用済み環境）
- [ ] `go test -race ./...`
- [ ] `staticcheck ./...`
- [ ] `gosec ./...`
- [ ] `risk-register.md` の High 項目レビュー
- [ ] `link-audit.md` の再確認（README/Index同期）

---

## 月次ルーチン
- [ ] `govulncheck ./...` 再実行
- [ ] launchd 運用確認（macOS）
- [ ] incident-playbook の更新要否判断
- [ ] quickstart / patch手順の妥当性見直し

---

## リリース前ルーチン
- [ ] minimal-regression-suite を実行
- [ ] acceptance-signoff-template を埋める
- [ ] final-status を更新

---

## 記録先
- 検証結果: `validation-log-template.md`
- 未解決課題: `risk-register.md`
- 重要変更履歴: `execution-changelog.md`
