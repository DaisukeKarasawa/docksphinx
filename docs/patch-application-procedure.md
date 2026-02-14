# Docksphinx MVP パッチ適用手順（実行マニュアル）

## 目的
`outputs/patches/*.diff.md` に定義した実装提案を、安全かつ再現可能に適用するための実行手順を提供する。

---

## 1. 事前確認

1. 現在ブランチ確認
   ```bash
   git branch --show-current
   ```
2. 作業ツリー確認
   ```bash
   git status --short
   ```
3. Go環境確認
   ```bash
   go version
   ```

---

## 2. 適用フェーズ

## フェーズ1-2適用
対象:
- `outputs/patches/phase1-2.diff.md`

適用後チェック:
```bash
make proto
go test ./...
```

成功条件:
- `cmd/docksphinxd` / `cmd/docksphinx` がビルド対象として成立
- `docksphinx/api/docksphinx/v1` 欠落エラーが解消

## フェーズ3-6適用
対象:
- `outputs/patches/phase3-6.diff.md`

適用後チェック:
```bash
go test ./...
go test -race ./...
govulncheck ./...
staticcheck ./...
gosec ./...
```

---

## 3. ロールバック戦略

- 原則: フェーズ単位でコミットを分離する
- フェーズ不具合時:
  - 該当フェーズコミットのみを `git revert` する
  - 前段フェーズの正常状態を維持する

---

## 4. ログ取得（推奨）

- テストログ:
  ```bash
  go test ./... 2>&1 | tee test.log
  ```
- raceログ:
  ```bash
  go test -race ./... 2>&1 | tee race.log
  ```
- セキュリティログ:
  ```bash
  govulncheck ./... 2>&1 | tee vuln.log
  staticcheck ./... 2>&1 | tee staticcheck.log
  gosec ./... 2>&1 | tee gosec.log
  ```

---

## 5. 既知の注意点

1. 現時点の既知課題
   - `govulncheck` が依存解決の内部エラーを返す場合がある
   - `gosec` が gRPC未生成/SSA要因で解析失敗することがある
2. 対応方針
   - まず phase1-2 適用で proto/cmd/api を揃える
   - 再度スキャンを実行して評価する

---

## 6. 関連ドキュメント
- `docs/requirements.md`
- `docs/mvp-acceptance-matrix.md`
- `docs/implementation-handoff-runbook.md`
- `docs/security-check-summary.md`
