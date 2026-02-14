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

---

## 2026-02-14 (status unknown pid checker diagnostics pass)

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

## 2026-02-14 (stop pid resolution extraction test pass)

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

## 2026-02-14 (stop wait cancel/timeout message separation pass)

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

## 2026-02-14 (resolveAddress precedence/config load test pass)

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

## 2026-02-14 (image created timestamp N/A rendering pass)

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

## 2026-02-14 (snapshot image created N/A output assertion pass)

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

## 2026-02-14 (localhost case-insensitive loopback validation pass)

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

## 2026-02-14 (client localhost case-insensitive warning suppression pass)

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

### Manual terminal E2E: uppercase localhost should not emit plaintext warning

```bash
./bin/docksphinx snapshot --addr LOCALHOST:65535
```

観測結果（抜粋）:
- `Error: connect daemon (LOCALHOST:65535): wait for grpc readiness LOCALHOST:65535: context deadline exceeded. start daemon with \`docksphinxd start\``

判定: PASS（`WARNING: connecting ... over plaintext` が出力されない）

---

## 2026-02-14 (warnInsecure uppercase localhost regression test pass)

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

## 2026-02-14 (snapshot resource section deterministic ordering pass)

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

---

## 2026-02-14 (recent events newest-first deterministic ordering pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx` の `TestSelectRecentEvents` で、入力順に依存せず `timestamp desc`（同値時 `id asc`）になることを検証。

---

## 2026-02-14 (gRPC snapshot compose group ordering stabilization pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc` の `TestStateToSnapshotSortsComposeGroupsAndFields` で、`StateToSnapshot` が `groups` と内部配列（ids/names/networks）を決定的順序に整列することを検証。

---

## 2026-02-14 (monitor compose group container-id ordering stabilization pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/monitor` の `TestBuildComposeGroupsUsesComposeLabels` で、group 内 `ContainerIDs` が昇順で安定化されることを検証。

---

## 2026-02-14 (recent-events tie-break regression assertion pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx` の `TestSelectRecentEvents` に、同一 timestamp での `id asc` タイブレーク保証ケースを追加。

---

## 2026-02-14 (event history mutation-isolation hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` に `history_test.go` を追加し、以下を検証:
  - `Add` 後に呼び出し側 `Event` を変更しても履歴が汚染されない
  - `Recent` の返却値を変更しても履歴本体が汚染されない
  - `maxSize` による上限保持と newest-first 返却順

---

## 2026-02-14 (recent-events nil filtering regression assertion pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx` の `TestSelectRecentEvents` に、`nil` イベントを含む入力でも `nil` を除去しつつ順序を維持して返却するケースを追加。

---

## 2026-02-14 (grpc event conversion contract test pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc` に以下の契約テストを追加:
  - `TestEventsToProtoSkipsNilAndConvertsFields`
  - `TestEventToProtoNil`

---

## 2026-02-14 (event history deep-copy hardening for nested data pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryAddAndRecentAreMutationSafe` を拡張し、`Data` 内のネストした map/slice 変更が履歴へ波及しないことを確認。

---

## 2026-02-14 (event history lock-hold reduction pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused change

- `internal/event.History` で `Add` の clone と `Recent` の deep-copy をロック外へ移し、ロック保持区間を最小化（挙動は維持）。

---

## 2026-02-14 (event history concurrent access regression test pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryConcurrentAddAndRecent` を追加し、`Add` と `Recent` の同時実行時にも上限契約（`len<=limit`）と非nil返却が維持されることを確認。

---

## 2026-02-14 (generic deep-copy coverage for typed event-data containers pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryAddAndRecentAreMutationSafe` を拡張し、`Data` に `map[string]string` / `[]string` を含むケースでも入力・出力ミューテーションが履歴に波及しないことを確認。

---

## 2026-02-14 (event history struct-payload deep-copy pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryAddAndRecentAreMutationSafe` で、`Data` に構造体（内部に `[]string` と `map[string]string` を保持）を含む場合でも、入力・出力ミューテーションが履歴へ波及しないことを確認。

---

## 2026-02-14 (event history pointer-payload deep-copy pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryAddAndRecentAreMutationSafe` を拡張し、`Data` に `*structuredPayload` を含む場合でも入力・返却値ミューテーションが履歴へ波及しないことを確認。

---

## 2026-02-14 (event data deep-copy map-key identity preservation pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryAddAndRecentAreMutationSafe` に `map[*int]string` ケースを追加し、deep copy 後もキー同一性（pointer key）が維持されることを確認。

---

## 2026-02-14 (event history boundary-limit contract test pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` に以下を追加:
  - `TestNewHistoryEnforcesMinSize`
  - `TestHistoryRecentLimitContract`

---

## 2026-02-14 (event history array-payload deep-copy pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryAddAndRecentAreMutationSafe` を拡張し、`Data` に `[2][]string`（配列内に参照型要素）を含むケースでも入力・返却値ミューテーションが履歴へ波及しないことを確認。

---

## 2026-02-14 (event history nil-safety contract test pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` に `TestHistoryNilSafetyContracts` を追加し、以下を固定:
  - `(*History)(nil).Add(...)` は panic せず no-op
  - `(*History)(nil).Recent(...)` は `nil` を返す
  - `History.Add(nil)` は履歴を汚染しない

---

## 2026-02-14 (event history typed-nil data preservation pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` に `TestHistoryPreservesTypedNilDataValues` を追加し、`Event.Data` 内の typed nil（`map[string]string(nil)`, `[]string(nil)`, `*structuredPayload(nil)`）が deep-copy 後も nil 性を保持することを確認。

---

## 2026-02-14 (event history nested typed-nil preservation pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` に `TestHistoryPreservesTypedNilInNestedContainers` を追加し、`Event.Data` のネストした `map`/`slice` 内でも typed nil が保持されることを確認。

---

## 2026-02-14 (event history limit-order and Add(nil) contract tightening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の契約テストを強化:
  - `TestHistoryRecentLimitContract` で `limit<=0` / `limit>len` の newest-first 順序を明示検証
  - `TestHistoryNilSafetyContracts` で `Add(nil)` が既存履歴を維持することを確認

---

## 2026-02-14 (selectRecentEvents boundary contract coverage pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx` の `TestSelectRecentEvents` を拡張し、以下を追加検証:
  - `limit<0` は `nil` を返す
  - `nil` 混在入力 + `limit` 制限時に、フィルタ後のソート結果から先頭 N 件が返る

---

## 2026-02-14 (selectRecentEvents non-mutating input pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx` の `TestSelectRecentEvents` を拡張し、入力スライス順序が関数呼び出し後も変化しない（non-mutating）こと、および全要素 `nil` 入力で `nil` が返ることを確認。

---

## 2026-02-14 (event history independent clone instance pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/event` の `TestHistoryAddAndRecentAreMutationSafe` を拡張し、`Recent()` の連続呼び出しで返却される `*Event` と `*structuredPayload` が同一ポインタ再利用ではなく、毎回独立 clone であることを確認。

---

## 2026-02-14 (selectRecentEvents alias-isolation pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx` の `selectRecentEvents` を `proto.Clone` ベースへ変更し、返却イベントのミューテーションが入力 snapshot の `Event` に波及しないことを確認。
- `TestSelectRecentEvents` に参照非共有（pointer inequality）と `Data` map の非波及を検証するケースを追加。

---

## 2026-02-14 (selectRecentEvents clone-scope optimization pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.selectRecentEvents` を最適化し、全候補 clone ではなく「ソート後の上位 `limit` 件のみ clone」へ変更。
- 既存の alias-isolation 回帰テスト（pointer inequality / data non-propagation）が引き続き PASS することを確認。

---

## 2026-02-14 (grpc StateToSnapshot non-mutating sorting regression pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.TestStateToSnapshotSortsComposeGroupsAndFields` を拡張し、`StateToSnapshot` 実行後も `StateManager` 側の `Groups` 順序および group 内スライス順序が未変更であることを確認。
- 併せて、呼び出し元入力スライス（`inputGroups`）が変更されないことを確認。

---

## 2026-02-14 (monitor compose grouping non-mutating regression pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/monitor.TestBuildComposeGroupsDoesNotMutateInputStateNetworks` を追加し、`buildComposeGroups` 呼び出し後も入力 `ContainerState.NetworkNames` が未変更であることを確認。

---

## 2026-02-14 (snapshot rendering non-mutating regression pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.TestPrintSnapshotToDoesNotMutateSnapshotOrderingFields` を追加し、`printSnapshotTo` 実行後も `Snapshot` の `Groups/Networks/Volumes/Images` の順序および group 内 `NetworkNames` 順序が入力のまま維持されることを確認。

---

## 2026-02-14 (snapshot rendering non-mutating containers/events regression pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.TestPrintSnapshotToDoesNotMutateSnapshotOrderingFields` を拡張し、`printSnapshotTo` 実行後も `Snapshot.Containers` および `Snapshot.RecentEvents` の入力順序が変化しないことを確認。

---

## 2026-02-14 (grpc resource sorting non-mutating regression pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.TestStateToSnapshotSortsResourcesWithoutMutatingSource` を追加し、`StateToSnapshot` が `Images/Networks/Volumes` を出力用にソートしても、呼び出し元入力および `StateManager` 内の保持順序を変更しないことを確認。

---

## 2026-02-14 (selectRecentEvents cross-call clone independence pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.TestSelectRecentEvents` を拡張し、`selectRecentEvents` の連続呼び出し結果が相互に参照共有せず（独立 clone）、1回目の返却値ミューテーションが2回目返却値へ波及しないことを確認。

---

## 2026-02-14 (tui detail sorting non-mutating regression pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.TestFilteredContainerRowsForDetailSortAndNonMutating` を追加し、`filteredContainerRowsForDetail` の sort mode（CPU/MEM/Uptime/Name）ごとの順序契約を固定。
- 併せて、同関数の実行で `Snapshot.Containers` の入力順序が変化しないこと（non-mutating）を確認。

---

## 2026-02-14 (tui image created N/A rendering pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.renderImages` の created 表示を `formatDateTimeOrNA` 化し、`CreatedUnix<=0` を `N/A` 表示へ統一。
- `TestFormatDateTimeOrNA` と `TestRenderImagesShowsNAForMissingCreatedTimestamp` を追加し、ヘルパー契約と TUI 実表示を固定。

---

## 2026-02-14 (tui container sort tie-break stabilization pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx` の `renderContainers` と `filteredContainerRowsForDetail` の `sortCPU/sortMemory/sortUptime` に、同値時 `ContainerName asc` タイブレークを追加し、表示順の揺れを抑制。
- `TestFilteredContainerRowsForDetailUsesNameTieBreakForStableOrdering` を追加し、CPU/MEM/Uptime 同値ケースで name tie-break が適用されることを確認。

---

## 2026-02-14 (snapshot/conversion deterministic tie-break hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.printSnapshotTo` のソートに同値時タイブレークを追加:
  - containers: `container_name asc` → tie `container_id asc`
  - networks: `name asc` → tie `driver/scope/network_id asc`
  - volumes: `name asc` → tie `driver/mountpoint/usage_note/ref_count asc`
  - images: `repository/tag asc` → tie `image_id asc`
  - groups: `project/service asc` → tie `container_ids/network_names` 連結値 asc
- `internal/grpc.StateToSnapshot` にも同等の tie-break を追加し、変換結果順の決定性を強化。
- 追加テスト:
  - `cmd/docksphinx.TestPrintSnapshotToUsesDeterministicTieBreakers`
  - `internal/grpc.TestStateToSnapshotUsesDeterministicTieBreakers`

---

## 2026-02-14 (tui same-name container-id tie-break hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx/tui.go` の `renderContainers` と `filteredContainerRowsForDetail` で、`ContainerName` 同値時に `ContainerId asc` を最終 tie-break として適用。
- `TestFilteredContainerRowsForDetailUsesContainerIDTieBreakWhenNamesEqual` を追加し、CPU/MEM/Uptime/Name の全 sort mode で同名コンテナ順が `container_id asc` になることを確認。

---

## 2026-02-14 (snapshot groups tie-break collision regression pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.TestPrintSnapshotToUsesDeterministicTieBreakers` を拡張し、`project/service` が同値な `GROUPS` 行でも `container_ids` ベース tie-break が適用され、表示順が決定的に固定されることを確認。

---

## 2026-02-14 (recent-events full tie-break ordering hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.selectRecentEvents` の比較関数を `lessRecentEvent` へ集約し、`timestamp desc` / `id asc` に加えて  
  `container_name` → `type` → `message` → `container_id` → `image_name` の tie-break を追加。
- `TestSelectRecentEvents` を拡張し、`timestamp` と `id` が同値（空）なケースでも追加 tie-break 連鎖により順序が固定されることを確認。

---

## 2026-02-14 (grpc events conversion deterministic ordering pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.EventsToProto` を copy-on-sort 化し、入力 slice を破壊せずに deterministic order で proto へ変換:
  - `timestamp desc`
  - `id asc`
  - `container_name asc`
  - `type asc`
  - `message asc`
  - `container_id asc`
  - `image_name asc`
- `TestEventsToProtoSortsDeterministicallyWithoutMutatingInput` を追加し、同 timestamp/id 衝突ケースでの順序固定と input non-mutation を確認。

---

## 2026-02-14 (tui resource panel deterministic ordering pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx/tui.go` の `renderImages` / `renderNetworks` / `renderVolumes` / `renderGroups` を copy-on-sort 化し、表示順の決定性を強化（入力 snapshot は非破壊）。
- 同値キー時の tie-break を追加:
  - images: `repository/tag` tie `image_id`
  - networks: `name` tie `driver/scope/network_id`
  - volumes: `name` tie `driver/mountpoint/usage_note/ref_count`
  - groups: `project/service` tie `container_ids/network_names`
- 追加テスト:
  - `TestRenderImagesUsesDeterministicTieBreakersAndNonMutating`
  - `TestRenderNetworksUsesDeterministicTieBreakersAndNonMutating`
  - `TestRenderVolumesUsesDeterministicTieBreakersAndNonMutating`
  - `TestRenderGroupsUsesDeterministicTieBreakersAndNonMutating`

---

## 2026-02-14 (grpc recent-events second-level timestamp contract alignment pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.lessInternalEvent` の timestamp 比較を `time.Time` 全精度から `Unix()`（秒）へ変更し、CLI 側の `pb.Event.TimestampUnix` 並び順契約と一致させた。
- `TestEventsToProtoSortsDeterministicallyWithoutMutatingInput` を拡張し、同一秒でナノ秒のみ異なるイベントでも `id asc` が優先されることを確認。

---

## 2026-02-14 (shared recent-event ordering comparator refactor pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/eventorder` パッケージを新設し、recent-event の tie-break 優先順位を共通化:
  - `timestamp(unix seconds) desc`
  - `id asc`
  - `container_name asc`
  - `type asc`
  - `message asc`
  - `container_id asc`
  - `image_name asc`
- `cmd/docksphinx.selectRecentEvents` は `eventorder.LessPB` を利用。
- `internal/grpc.EventsToProto` は `eventorder.LessInternal` を利用。
- 追加テスト:
  - `internal/eventorder.TestLessPBAndLessInternalProduceSameOrder`
  - `internal/eventorder.TestLessInternalUsesSecondLevelTimestampBeforeID`

---

## 2026-02-14 (snapshot resource comparator centralization pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/snapshotorder` パッケージを新設し、以下の比較ロジックを共通化:
  - `LessContainerInfo`
  - `LessComposeGroup`
  - `LessNetworkInfo`
  - `LessVolumeInfo`
  - `LessImageInfo`
- `cmd/docksphinx/main.go`、`cmd/docksphinx/tui.go`、`internal/grpc/convert.go` の snapshot リソースソートを共通 comparator 利用に置換。
- 追加テスト:
  - `internal/snapshotorder.TestLessContainerInfo`
  - `internal/snapshotorder.TestLessComposeGroup`
  - `internal/snapshotorder.TestLessNetworkInfo`
  - `internal/snapshotorder.TestLessVolumeInfo`
  - `internal/snapshotorder.TestLessImageInfo`

---

## 2026-02-14 (tui container comparator deduplication pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx/tui.go` の `renderContainers` と `filteredContainerRowsForDetail` で重複していた sort predicate を `lessContainerForMode` + `lessContainerNameID` に集約。
- CPU/MEM/Uptime/Name の tie-break 契約（同値時 `container_name` → `container_id`）を単一実装へ統一し、表示/詳細でのドリフト余地を排除。
- 既存の TUI ソート回帰テスト群（`TestFilteredContainerRowsForDetail*`）が通過することを確認し、挙動不変を検証。

---

## 2026-02-14 (shared event comparator nil-safety hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/eventorder.LessPB` と `LessInternal` に nil-safe guard を追加し、比較時に `nil` が混在しても panic しないようにした（`non-nil < nil`）。
- 追加テスト:
  - `TestLessPBNilSafety`
  - `TestLessInternalNilSafety`

---

## 2026-02-14 (snapshot comparator nil-safety hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/snapshotorder` の比較関数（containers/groups/networks/volumes/images）に nil-safe guard を追加し、`nil` 混在比較で panic しないよう防御化（`non-nil < nil`）。
- 追加テスト:
  - `TestLessContainerInfoNilSafety`
  - `TestLessComposeGroupNilSafety`
  - `TestLessNetworkInfoNilSafety`
  - `TestLessVolumeInfoNilSafety`
  - `TestLessImageInfoNilSafety`

---

## 2026-02-14 (compose-group key collision hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/monitor.buildComposeGroups` の集約キーを `project + "|" + service` 文字列連結から `struct{project, service}` へ変更し、区切り文字を含む値での衝突マージを排除。
- `TestBuildComposeGroupsProjectServiceDelimiterCollisionSafety` を追加し、
  - `project="a|b", service="c"`
  - `project="a", service="b|c"`
  の2組が誤って1グループに融合しないことを確認。

---

## 2026-02-14 (compose-group comparator canonicalization pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/snapshotorder.LessComposeGroup` の tie-break 比較で `container_ids` / `network_names` / `container_names` を比較前にコピーしてソートし、入力配列順のゆらぎに依存しない canonical order へ強化。
- `TestLessComposeGroupCanonicalizesSlicesAndKeepsInputsUnchanged` を追加し、以下を確認:
  - 比較時に内部スライス順を正規化して順序判定されること
  - 比較処理が元の `ComposeGroup` スライスを破壊しないこと（non-mutating）

---

## 2026-02-14 (cli/tui nil-entry render skip hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx.printSnapshotTo` の各セクション（containers/recent events/groups/networks/volumes/images）で `nil` 要素を明示スキップし、空行/空リソース行アーティファクトを抑止。
- `cmd/docksphinx/tui.go` の `renderContainers` / `renderImages` / `renderNetworks` / `renderVolumes` / `renderGroups` / `filteredContainerRowsForDetail` で `nil` 要素をスキップ。
- 追加テスト:
  - `TestPrintSnapshotToSkipsNilResourceEntries`
  - `TestRenderResourcesSkipNilEntries`
  - `TestFilteredContainerRowsForDetailSkipsNilEntries`

---

## 2026-02-14 (tui event buffer nil-filter compaction pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `cmd/docksphinx/tui.go` に `compactEvents` を追加し、stream受信イベント配列で `nil` を除去しつつ上限件数を一元管理。
- `consumeStream` の snapshot/event 更新時に `compactEvents` を利用し、`m.events` 内の `nil` 混入と過剰保持を抑止。
- `renderRight` / `lastEventType` で `nil` イベントを明示スキップし、右ペインの空行アーティファクトと誤判定を防止。
- 追加テスト:
  - `TestCompactEventsFiltersNilAndAppliesLimit`
  - `TestLastEventTypeSkipsNilEntries`

---

## 2026-02-14 (grpc stream initial-send error propagation hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.Server.Stream` で `IncludeInitialSnapshot=true` 時の初回 `stream.Send(snapshot)` 失敗を握りつぶさず、そのまま呼び出し元へ返すよう修正。
- イベント配信ループで `EventToProto(ev)==nil` を防御的にスキップし、`nil` payload の送信を抑止。
- 追加テスト:
  - `TestServerStreamReturnsInitialSnapshotSendError`
  - `TestServerStreamSkipsNilEventPayloads`

---

## 2026-02-14 (grpc server nil-dependency guard hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.Server.GetSnapshot` に `engine==nil` ガードを追加し、panic ではなく `codes.Unavailable` を返すよう修正。
- `internal/grpc.Server.Stream` に `engine==nil` / `bcast==nil` ガードを追加し、依存欠落時に `codes.Unavailable` を返すよう修正。
- `recentEventLimit` ヘルパーを追加し、`opts==nil` または `RecentEventLimit<=0` でも既定値（50）で安全に処理。
- 追加テスト:
  - `TestServerGetSnapshotReturnsUnavailableWhenEngineMissing`
  - `TestServerStreamReturnsUnavailableWhenDependenciesMissing`

---

## 2026-02-14 (grpc stream nil-argument guard hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.Server.Stream` 冒頭で `stream==nil` を `codes.InvalidArgument` として明示拒否し、`stream.Context()` 参照時 panic を予防。
- `TestServerStreamReturnsUnavailableWhenDependenciesMissing` を拡張し、`engine/bcast` が揃っていても `stream=nil` の場合は `InvalidArgument` が返ることを回帰固定。

---

## 2026-02-14 (grpc handler early context-cancel hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.Server.GetSnapshot` で `ctx.Err()` を先頭評価し、キャンセル済み context では依存チェックより先に `status.FromContextError(err)` を返すよう修正。
- `internal/grpc.Server.Stream` で `stream.Context().Err()` を先頭評価し、既に終了済み stream context では同様に context 由来 status を返すよう修正。
- 追加テスト:
  - `TestServerGetSnapshotReturnsContextErrorWhenCanceled`
  - `TestServerStreamReturnsContextErrorWhenAlreadyCanceled`

---

## 2026-02-14 (grpc server nil-receiver and uninitialized-start hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.Server.Start` で未初期化状態（`s==nil` / `grpc==nil` / `lis==nil`）を明示検出し、`Serve` 呼び出し前にエラーを返すよう修正。
- `internal/grpc.Server.Stop` で `s==nil` を no-op として扱い、nil receiver 呼び出しでも panic しないよう修正。
- `internal/grpc.Server.GetSnapshot` / `Stream` で `s==nil` を `codes.Unavailable` として返し、nil receiver 呼び出し時 panic を防止。
- 追加テスト:
  - `TestServerStartReturnsErrorWhenUninitialized`
  - `TestServerStopNilReceiverNoPanic`
  - `TestServerGetSnapshotReturnsUnavailableWhenReceiverNil`
  - `TestServerStreamReturnsUnavailableWhenReceiverNil`

---

## 2026-02-14 (grpc client boundary hardening pass)

### Unified gate run

```bash
go test ./...
make quality
```

結果:
- `go test ./...`: PASS
- `make quality`: PASS
  - `make test`: PASS
  - `make test-race`: PASS
  - `make security`: PASS
    - `gosec`: PASS (Issues: 0)
    - `govulncheck -mode=binary`: PASS
    - `govulncheck ./...`: known internal error (warning)

### Focused regression assertion

- `internal/grpc.NewClient` / `Client.GetSnapshot` / `Client.Stream` で `ctx==nil` を `context.Background()` へ正規化し、`nil` context 起因の panic リスクを防止。
- `internal/grpc.waitUntilReady` に `conn==nil` ガードを追加し、未初期化接続での panic を防止。
- `internal/grpc/client_test.go` を新規追加し、以下の契約を回帰固定:
  - `TestClientGetSnapshotAndStreamForwardContextAndRequests`
  - `TestClientMethodsReturnErrorWhenClientIsNil`
  - `TestWaitUntilReadyRejectsNilConnection`
  - `TestNewClientRejectsEmptyAddress`
