# macOS launchd 導入チェックリスト（MVP）

## 目的
`docksphinxd` の常駐運用を launchd で有効化する際の確認漏れを防ぐ。

---

## 1. 事前準備
- [ ] plist を `~/Library/LaunchAgents/io.docksphinx.docksphinxd.plist` に配置
- [ ] `ProgramArguments` のパスが正しい
- [ ] `--config` のパスが存在する
- [ ] ログ出力先ディレクトリが存在する

---

## 2. 起動確認
- [ ] `launchctl bootstrap gui/$(id -u) ...` 実行済み
- [ ] `launchctl print gui/$(id -u)/io.docksphinx.docksphinxd` でロード済み
- [ ] `docksphinx status` で稼働確認

---

## 3. 機能確認
- [ ] `docksphinx snapshot` が応答する
- [ ] `docksphinx tail` が受信する
- [ ] `docksphinx tui` が接続する

---

## 4. 復帰性確認
- [ ] ログアウト/ログイン後に自動起動する
- [ ] プロセス終了後に再起動する（KeepAlive）

---

## 5. 停止確認
- [ ] `launchctl bootout ...` で停止できる
- [ ] `docksphinx stop` でも停止できる

---

## 6. トラブル時
- [ ] `~/Library/Logs/docksphinx/*.log` を確認
- [ ] `docs/incident-playbook.md` の該当手順を実施
