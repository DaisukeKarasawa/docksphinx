# Docksphinx MVP コマンドリファレンス（適用後想定）

## 目的
実装パッチ適用後に使う主要コマンドを1ページで参照できるようにする。

---

## 1. ビルドと生成

```bash
make proto
make build
```

---

## 2. daemon

起動:
```bash
./bin/docksphinxd --config configs/docksphinx.yaml.example
```

状態確認:
```bash
./bin/docksphinx status
```

停止:
```bash
./bin/docksphinx stop
```

---

## 3. CLI監視

スナップショット:
```bash
./bin/docksphinx snapshot --addr 127.0.0.1:50051
```

tail:
```bash
./bin/docksphinx tail --addr 127.0.0.1:50051
```

TUI:
```bash
./bin/docksphinx tui --addr 127.0.0.1:50051
```

---

## 4. テスト

```bash
go test ./...
go test -race ./...
```

---

## 5. セキュリティ/静的解析

```bash
govulncheck ./...
staticcheck ./...
gosec ./...
```

---

## 6. launchd（macOS）

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/io.docksphinx.docksphinxd.plist
launchctl print gui/$(id -u)/io.docksphinx.docksphinxd
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/io.docksphinx.docksphinxd.plist
```
