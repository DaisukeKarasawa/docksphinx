# Docksphinx MVP 検証ログテンプレート

## 目的
適用後の検証結果を、誰が実行しても同じ形式で記録できるようにする。

---

## 実行環境
- Date:
- OS:
- Go version:
- Branch:
- Commit:

---

## 1. Build / Generate

### コマンド
```bash
make proto
make build
```

### 結果
- [ ] Success
- [ ] Failure
- メモ:

---

## 2. Unit / Integration

### コマンド
```bash
go test ./...
go test -race ./...
```

### 結果
- `go test ./...`:
- `go test -race ./...`:
- 失敗時ログ:

---

## 3. Runtime Smoke

### コマンド
```bash
./bin/docksphinxd --config configs/docksphinx.yaml.example
./bin/docksphinx snapshot --addr 127.0.0.1:50051
./bin/docksphinx tail --addr 127.0.0.1:50051
./bin/docksphinx tui --addr 127.0.0.1:50051
```

### 結果
- daemon起動:
- snapshot:
- tail:
- tui:

---

## 4. Security / Static Analysis

### コマンド
```bash
govulncheck ./...
staticcheck ./...
gosec ./...
```

### 結果
- govulncheck:
- staticcheck:
- gosec:

---

## 5. 受け入れ判定
- [ ] A daemon
- [ ] B config
- [ ] C log/history
- [ ] D snapshot/tail stability
- [ ] E TUI ops
- [ ] F resource collection
- [ ] G uptime/volume policy
- [ ] H1 launchd
- [ ] H2 backpressure/leak
- [ ] H3 cooldown
- [ ] H4 compose grouping
- [ ] H5 metric definition
- [ ] H6 test strategy

最終判定:
- [ ] PASS
- [ ] CONDITIONAL PASS
- [ ] FAIL

備考:
