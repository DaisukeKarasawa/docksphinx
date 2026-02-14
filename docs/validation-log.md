# Validation Log (MVP)

このログは、MVP実装に対して実際に実行した主要検証コマンドと結果を記録します。

## 2026-02-14

### Build

```bash
make build
```

結果: PASS

### Unit / Integration (local)

```bash
go test ./...
go test -race ./...
```

結果: PASS

### Focused tests

```bash
go test -v ./internal/grpc -run TestServerSnapshotAndStreamInitial
go test -v ./internal/monitor -run TestThresholdMonitorCooldown
```

結果: PASS

### Static / Security checks

```bash
"$(go env GOPATH)/bin/staticcheck" ./...
"$(go env GOPATH)/bin/gosec" -exclude-dir=api ./...
"$(go env GOPATH)/bin/govulncheck" ./...
```

- staticcheck: PASS
- gosec: PASS (Issues: 0, excluding generated `api/`)
- govulncheck: FAIL (tool internal error)

govulncheck エラー:

```text
internal error: package "golang.org/x/text/encoding" without types was imported from "github.com/gdamore/encoding"
```

備考:
- `go mod tidy`、`go list -deps ./...`、`govulncheck` バージョン固定、`GOTOOLCHAIN` 固定でも再現。
- `make security` はこの internal error を warning として表示し、他ゲート結果は有効とする。

### Runtime smoke (environment-limited)

```bash
./bin/docksphinxd start --config ./configs/docksphinx.yaml.example
./bin/docksphinxd status --config ./configs/docksphinx.yaml.example
./bin/docksphinx snapshot --config ./configs/docksphinx.yaml.example
timeout 5s ./bin/docksphinx tail --config ./configs/docksphinx.yaml.example
```

結果:
- 実行環境に Docker daemon/CLI が存在しないため、daemon 起動は期待どおり安全失敗。
- status/snapshot/tail のエラー経路は健全動作を確認。
