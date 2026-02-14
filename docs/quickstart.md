# Docksphinx MVP Quickstart

## 目的
初回参加者が最短で「実装提案の適用」と「検証」まで進めるための手順。

---

## 0. 先に読む
1. `docs/requirements.md`
2. `docs/mvp-acceptance-matrix.md`
3. `docs/requirements-traceability.md`
4. `docs/final-status.md`

---

## 1. パッチ適用（順序厳守）
1. `outputs/patches/phase1-2.diff.md` を適用
2. `make proto`
3. `go test ./...`
4. `outputs/patches/phase3-6.diff.md` を適用
5. `go test ./...`
6. `go test -race ./...`

---

## 2. 実行確認

```bash
./bin/docksphinxd --config configs/docksphinx.yaml.example
./bin/docksphinx snapshot --addr 127.0.0.1:50051
./bin/docksphinx tail --addr 127.0.0.1:50051
./bin/docksphinx tui --addr 127.0.0.1:50051
```

---

## 3. セキュリティ/静的解析

```bash
govulncheck ./...
staticcheck ./...
gosec ./...
```

---

## 4. 失敗時の参照先
- `docs/faq.md`
- `docs/security-check-summary.md`
- `docs/limitations-and-next-steps.md`
- `docs/risk-register.md`

---

## 5. 記録テンプレート
- `docs/validation-log-template.md` をコピーして実行ログを残す。

---

## 6. 最終判定
1. `docs/master-checklist.md` を確認
2. `docs/review-sequence.md` の順でレビュー
3. `docs/acceptance-signoff-template.md` に判定を記録
