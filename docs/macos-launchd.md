# macOS launchd での `docksphinxd` 常駐運用（MVP）

## 1. 目的
`docksphinxd` をログイン時に自動起動し、プロセス終了時に自動復帰できるようにします。  
MVPでは launchd 連携を「雛形 plist + 導入手順」で提供します。

## 2. 前提
- `docksphinxd` バイナリがインストール済み
- 設定ファイルが存在（例: `~/.config/docksphinx/config.yaml`）
- `daemon.pid_file` と `log.file` は絶対パス

## 3. plist の配置
1. `configs/io.docksphinx.docksphinxd.plist.example` をコピー:
   ```bash
   cp configs/io.docksphinx.docksphinxd.plist.example ~/Library/LaunchAgents/io.docksphinx.docksphinxd.plist
   ```
2. 以下を実環境のパスへ編集:
   - `ProgramArguments[0]` (`docksphinxd` の実パス)
   - `ProgramArguments` の `--config` パス
   - `StandardOutPath`, `StandardErrorPath`
   - `WorkingDirectory`

## 4. ログディレクトリ作成
```bash
mkdir -p ~/Library/Logs/docksphinx
```

## 5. 起動（登録）
```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/io.docksphinx.docksphinxd.plist
```

## 6. 状態確認
launchd 側:
```bash
launchctl print gui/$(id -u)/io.docksphinx.docksphinxd
```

daemon 側:
```bash
docksphinxd status --config ~/.config/docksphinx/config.yaml
docksphinx snapshot --config ~/.config/docksphinx/config.yaml
```

## 7. 停止（解除）
```bash
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/io.docksphinx.docksphinxd.plist
```

必要に応じて手動停止:
```bash
docksphinxd stop --config ~/.config/docksphinx/config.yaml
```

## 8. 運用メモ
- `KeepAlive=true` のため異常終了時に再起動されます。
- daemon 起動確認は **PIDファイル + gRPC疎通** の両方を使って判定します。
- gRPC は既定で `127.0.0.1` bind（外部公開しない）。
