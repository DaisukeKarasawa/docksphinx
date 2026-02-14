# Minimal Regression Suite（MVP）

## 目的
実装適用後に最低限回すべき回帰確認を固定し、検証漏れを防ぐ。

---

## 1. Build/Generate
- [ ] `make proto`
- [ ] `make build`

---

## 2. Automated
- [ ] `go test ./...`
- [ ] `go test -race ./...`

---

## 3. Runtime smoke
- [ ] `docksphinxd` 起動
- [ ] `snapshot` 応答
- [ ] `tail` 受信継続
- [ ] `tui` 接続と基本操作（Tab/jk/検索/ソート/pause/q）

---

## 4. Security/Static
- [ ] `govulncheck ./...`
- [ ] `staticcheck ./...`
- [ ] `gosec ./...`

---

## 5. Must-pass判定
- [ ] daemonが安定稼働
- [ ] `go test ./...` が通る
- [ ] 重大セキュリティ指摘（High）が未解決で残っていない

---

## 6. 結果記録
- [ ] `docs/validation-log-template.md` に記録
- [ ] 未解決事項は `docs/risk-register.md` へ反映
