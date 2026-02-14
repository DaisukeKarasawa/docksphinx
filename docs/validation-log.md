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
"$(go env GOPATH)/bin/govulncheck" -mode=binary ./bin/docksphinx
"$(go env GOPATH)/bin/govulncheck" -mode=binary ./bin/docksphinxd
"$(go env GOPATH)/bin/govulncheck" ./...
```

- staticcheck: PASS
- gosec: PASS (Issues: 0, excluding generated `api/`)
- govulncheck (binary mode): PASS (No vulnerabilities found)
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

---

## 2026-02-14 (additional hardening pass)

### Re-run gates after daemon stop + TUI reconnect + grouping tests

```bash
make test-race
make security
```

結果:
- `go test -race ./...` : PASS
- `staticcheck` : PASS
- `gosec -exclude-dir=api ./...` : PASS (Issues: 0)
- `govulncheck -mode=binary` : PASS
- `govulncheck ./...` : internal error（既知、warning扱い）

### Toolchain security patch verification

- `go.mod` toolchain を `go1.24.13` へ更新。
- 更新前に binary-mode `govulncheck` で検出された標準ライブラリ脆弱性（GO-2026-4341 / GO-2026-4340 / GO-2026-4337）について、
  更新後の binary-mode 再スキャンで `No vulnerabilities found` を確認。

---

## 2026-02-14 (final stability pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`:
  - `staticcheck`: PASS
  - `gosec -exclude-dir=api`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS (`docksphinx`, `docksphinxd` ともに No vulnerabilities found)
  - `govulncheck ./...`: internal error（既知、warning表示）

---

## 2026-02-14 (post clone-isolation hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS（source-mode govulncheck は既知 internal error を warning 表示）

### Additional security hardening

- `grpc.allow_non_loopback=false`（既定）時に、`grpc.address` が loopback 以外なら設定バリデーションで拒否することを追加。
- `internal/config` テストで以下を確認:
  - 既定では `0.0.0.0:50051` を拒否
  - `allow_non_loopback=true` では許可
