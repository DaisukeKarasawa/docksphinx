# Docksphinx MVP FAQ

## Q1. なぜコードを直接変更していないのですか？
A. リポジトリ運用ルールで直接変更が禁止されているため、実装提案は `outputs/` で提供している。

## Q2. `go test ./...` が失敗する主因は？
A. `docksphinx/api/docksphinx/v1`（proto生成物）が未存在なため。  
まず `phase1-2` の適用と `make proto` が必要。

## Q3. volume usage が `N/A (metadata-only)` なのは不具合ですか？
A. 不具合ではない。MVPでは Docker API制約を踏まえて metadata-based 代替指標を採用している。

## Q4. `govulncheck` が内部エラーで落ちる場合は？
A. Goバージョン/依存解決との相性問題がある。  
`make proto` 後に再実行し、必要なら toolchain を固定して再試行する。

## Q5. `gosec` の G115 はどう扱うべきですか？
A. `uint64 -> int64` 変換の overflow 懸念。  
型統一または安全変換関数の導入を優先して対応する。

## Q6. TUIの最低操作要件は？
A. Tab/矢印/jk/`/`検索/ソート切替/pause/q。

## Q7. launchd の状態確認は？
A. 以下で確認:
```bash
launchctl print gui/$(id -u)/io.docksphinx.docksphinxd
```

## Q8. どの順で成果物を読むべき？
A.
1. `docs/requirements.md`
2. `docs/mvp-acceptance-matrix.md`
3. `docs/patch-application-procedure.md`
4. `outputs/patches/phase1-2.diff.md`
5. `outputs/patches/phase3-6.diff.md`
