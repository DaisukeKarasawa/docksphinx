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

---

## 2026-02-14 (start precondition hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (PID fail-fast hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (tail EOF reconnect hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (status fail-safe hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (tail retry logging hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

### Manual terminal E2E: tail retry stderr message

```bash
./bin/docksphinx tail --addr 127.0.0.1:65535
```

観測結果（抜粋）:
- `tail connect failed: wait for grpc readiness 127.0.0.1:65535: context deadline exceeded (retrying in 500ms)`
- `tail connect failed: wait for grpc readiness 127.0.0.1:65535: context deadline exceeded (retrying in 1s)`

判定: PASS（接続失敗理由と再試行待機時間を stderr に表示）

---

## 2026-02-14 (tail retry log helper test pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (CLI daemon guidance error-path test pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (CLI plaintext warning boundary test pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (direct ECONNREFUSED detection hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

### Manual terminal E2E: snapshot error guidance

```bash
./bin/docksphinx snapshot --addr 127.0.0.1:65535
```

観測結果（抜粋）:
- `Error: connect daemon (127.0.0.1:65535): wait for grpc readiness 127.0.0.1:65535: context deadline exceeded. start daemon with \`docksphinxd start\``

判定: PASS（接続失敗時に起動ガイダンスを表示）

---

## 2026-02-14 (tail stream reconnect wording hardening pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (status duplicate error-line suppression pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

### Manual terminal E2E: status output on daemon-down

```bash
./bin/docksphinxd status
```

観測結果（抜粋）:
- `status: not running (pid: not found, grpc=127.0.0.1:50051, err=dial daemon: wait for grpc readiness 127.0.0.1:50051: context deadline exceeded)`

判定: PASS（`Error:` 行の重複出力なし、終了コード1は維持）

---

## 2026-02-14 (idempotent stop on missing pid-file pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

### Manual terminal E2E: stop without pid-file

```bash
./bin/docksphinxd stop; echo EXIT:$?
```

観測結果（抜粋）:
- `Daemon is already stopped (pid file not found)`
- `EXIT:0`

判定: PASS（PIDファイル未存在時に成功終了し、stop が冪等）

---

## 2026-02-14 (health check timeout context propagation pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

---

## 2026-02-14 (preserve original error with already-reported marker pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

### Manual terminal E2E: status output invariant after marker change

```bash
./bin/docksphinxd status
```

観測結果（抜粋）:
- `status: not running (pid: not found, grpc=127.0.0.1:50051, err=dial daemon: wait for grpc readiness 127.0.0.1:50051: context deadline exceeded)`

判定: PASS（`Error:` 行の重複出力なしを維持）

---

## 2026-02-14 (readPID contract test expansion pass)

### Unified gate run

```bash
make quality
```

結果:
- `make test`: PASS
- `make test-race`: PASS
- `make security`: PASS
  - `gosec`: PASS (Issues: 0)
  - `govulncheck -mode=binary`: PASS
  - `govulncheck ./...`: known internal error (warning)

### Additional security hardening

- `grpc.allow_non_loopback=false`（既定）時に、`grpc.address` が loopback 以外なら設定バリデーションで拒否することを追加。
- `internal/config` テストで以下を確認:
  - 既定では `0.0.0.0:50051` を拒否
  - `allow_non_loopback=true` では許可
