# Docksphinx

Docker 環境をローカルで継続監視し、`snapshot` / `tail` / `tui` で可視化する MVP ツールです。

## Build

```bash
make build
```

## Quality Gates

```bash
make test
make test-race
make security
```

`make security` は `staticcheck` / `gosec` / `govulncheck` を実行します。  
`govulncheck` は環境依存で内部エラーになる場合があり、その場合は warning を表示して継続します。  
internal error 以外の失敗は `make security` をエラー終了します。  
詳細は `docs/security-check-summary.md` を参照してください。

## Daemon

起動:
```bash
./bin/docksphinxd start --config ./configs/docksphinx.yaml.example
```

状態確認:
```bash
./bin/docksphinxd status --config ./configs/docksphinx.yaml.example
```

停止:
```bash
./bin/docksphinxd stop --config ./configs/docksphinx.yaml.example
```
`stop` は SIGTERM 送信後、最大5秒プロセス終了を待機して結果を返します。
既に停止済みで PID が stale の場合は、PID ファイルを自動で削除します。

## CLI

スナップショット:
```bash
./bin/docksphinx snapshot --config ./configs/docksphinx.yaml.example
```
コンテナ一覧に加えて、直近イベント（最新10件）を表示します。

ストリーム:
```bash
./bin/docksphinx tail --config ./configs/docksphinx.yaml.example
```

TUI:
```bash
./bin/docksphinx tui --config ./configs/docksphinx.yaml.example
```

## TUI キー操作

- `Tab` / `←` / `→`: パネル切替
- `j` / `k` または `↑` / `↓`: 移動
- `/`: 検索フィルタ
- `s`: ソート切替（containers）
- `p`: 一時停止（表示は継続）
- `q`: 終了

## 重要な MVP 制約

- 読み取り専用（コンテナ実行操作なし）
- gRPC は既定で `127.0.0.1` bind
- volume 使用量は Docker API 制約により **metadata-only**

詳細は以下を参照:
- `docs/metrics-definition.md`
- `docs/macos-launchd.md`
