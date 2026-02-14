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
make quality
```

`make security` は `staticcheck` / `gosec` / `govulncheck` を実行します。  
`govulncheck` はまず binary mode で実行ファイルを検査します。  
`govulncheck` は環境依存で内部エラーになる場合があり、その場合は warning を表示して継続します。  
internal error 以外の失敗は `make security` をエラー終了します。  
詳細は `docs/security-check-summary.md` を参照してください。

## Daemon

起動:
```bash
./bin/docksphinxd start --config ./configs/docksphinx.yaml.example
```
`start` は既存 PID が稼働中なら二重起動を防止し、stale PID の場合は自動掃除して起動します。
PID ファイルが破損している場合は誤動作防止のためエラー終了します。

状態確認:
```bash
./bin/docksphinxd status --config ./configs/docksphinx.yaml.example
```
`status` は stale PID を検知した場合、PID ファイルを自動で掃除します。
PID ファイルが破損している場合は `status` もエラー終了します（fail-safe）。

停止:
```bash
./bin/docksphinxd stop --config ./configs/docksphinx.yaml.example
```
`stop` は SIGTERM 送信後、最大5秒プロセス終了を待機して結果を返します。
既に停止済みで PID が stale の場合は、PID ファイルを自動で削除します。
PID ファイルが存在しない場合も「既に停止済み」として成功終了します（冪等）。

## CLI

スナップショット:
```bash
./bin/docksphinx snapshot --config ./configs/docksphinx.yaml.example
```
コンテナ一覧に加えて、直近イベント（最新10件）を表示します。
さらに groups / networks / volumes / images の現在情報も表示します。
image の作成日時が欠損している場合は `N/A` を表示します。

ストリーム:
```bash
./bin/docksphinx tail --config ./configs/docksphinx.yaml.example
```
接続断（EOF/一時エラー）時は自動で再接続します（Ctrl+C で終了）。
再接続中は接続失敗理由と再試行待機時間を stderr に表示します。

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
- `grpc.allow_non_loopback=false`（既定）では loopback 以外の bind を拒否
- volume 使用量は Docker API 制約により **metadata-only**

詳細は以下を参照:
- `docs/metrics-definition.md`
- `docs/macos-launchd.md`

## MVP 完了チェック（実装済み）

### 必須要件（A〜G）

- A. `docksphinxd`（start/stop/status）: ✅
- B. `docksphinx`（snapshot/tail/tui）: ✅
- C. event history（in-memory ring）: ✅
- D. tail 再接続・backoff・停止規律: ✅
- E. TUI 4-pane + 必須キー操作: ✅
- F. containers/images/networks/volumes/groups の表示: ✅
- G. uptime / 欠損値 `N/A` 表示規約: ✅

### 高難度要件（H1〜H6）

- H1. launchd 運用ドキュメント: ✅ (`docs/macos-launchd.md`)
- H2. stream backpressure（bounded + drop policy）: ✅
- H3. threshold noise 抑止（cooldown）: ✅
- H4. compose grouping（labels優先 + network fallback）: ✅
- H5. metrics 定義の明文化: ✅ (`docs/metrics-definition.md`)
- H6. テスト戦略実施（unit/race/security/manual証跡）: ✅ (`docs/validation-log.md`)

### 既知の制約（2026-02-14時点）

- `govulncheck ./...` は環境依存の internal error を起こす場合があります。
  - 対策として binary mode (`-mode=binary`) を必須実行し、source mode internal error は warning 扱い。
  - 詳細: `docs/security-check-summary.md`
