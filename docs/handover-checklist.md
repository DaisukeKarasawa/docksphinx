# Docksphinx MVP ハンドオーバーチェックリスト

## 目的
`outputs` の実装提案を引き継ぐ際に、漏れなく実施できるようにする。

---

## 1. 受領確認
- [ ] `docs/requirements.md` を確認した
- [ ] `docs/mvp-acceptance-matrix.md` を確認した
- [ ] `docs/final-status.md` を確認した
- [ ] `docs/risk-register.md` を確認した

---

## 2. 適用準備
- [ ] ブランチが正しい（feature branch）
- [ ] 作業ツリーがクリーン
- [ ] Goツールチェーン確認済み
- [ ] Docker動作確認済み

---

## 3. 適用実行
- [ ] `outputs/patches/phase1-2.diff.md` を適用
- [ ] `make proto` 実行
- [ ] `go test ./...` 実行
- [ ] `outputs/patches/phase3-6.diff.md` を適用
- [ ] `go test ./...` / `go test -race ./...` 実行

---

## 4. セキュリティ/静的解析
- [ ] `govulncheck ./...` 実行
- [ ] `staticcheck ./...` 実行
- [ ] `gosec ./...` 実行
- [ ] 結果を `docs/validation-log-template.md` 形式で記録

---

## 5. 運用確認
- [ ] daemon起動/停止/状態確認
- [ ] snapshot/tail/tui 実動作
- [ ] launchd 手順（macOS）確認

---

## 6. ハンドオーバー完了条件
- [ ] 受け入れ基準 A〜G / H1〜H6 を満たす
- [ ] 未解決リスクを `docs/risk-register.md` に反映した
- [ ] 実行ログを保管した
