# コマンド出力サンプル（実行証跡）

## 目的
実際に実行したコマンドの出力例を保存し、検証結果の再現性と監査性を高める。

---

## 1) `go test ./...`

```text
# docksphinx/internal/grpc
internal/grpc/convert.go:8:2: package docksphinx/api/docksphinx/v1 is not in std
FAIL	docksphinx/internal/grpc [setup failed]
?   	docksphinx/internal/docker	[no test files]
?   	docksphinx/internal/event	[no test files]
ok  	docksphinx/internal/monitor	(cached)
FAIL
```

要点:
- 全体テスト失敗の一次原因は gRPC生成物欠落。

---

## 2) `go test ./internal/docker ./internal/event ./internal/monitor`

```text
?   	docksphinx/internal/docker	[no test files]
?   	docksphinx/internal/event	[no test files]
ok  	docksphinx/internal/monitor	(cached)
```

要点:
- 実行可能範囲のテストは成功。

---

## 3) `go test -race ./internal/docker ./internal/event ./internal/monitor`

```text
?   	docksphinx/internal/docker	[no test files]
?   	docksphinx/internal/event	[no test files]
ok  	docksphinx/internal/monitor	1.011s
```

要点:
- raceありでも monitor 範囲は成功。

---

## 4) `staticcheck ./...`（抜粋）

```text
internal/docker/errors.go:13:25: error strings should not be capitalized (ST1005)
internal/docker/errors.go:28:5: client.IsErrNotFound is deprecated (SA1019)
internal/monitor/engine_test.go:128:2: this value of events is never used (SA4006)
internal/grpc/convert.go:8:2: package docksphinx/api/docksphinx/v1 is not in std (compile)
```

要点:
- style/deprecation/unused + proto欠落に起因するcompile問題を検出。

---

## 5) `gosec ./...`（抜粋）

```text
G115: integer overflow conversion uint64 -> int64 (internal/docker/metrics.go)
```

要点:
- 数値型変換（G115）は優先修正候補。

---

## 6) `govulncheck ./...`（抜粋）

```text
internal error: package "...otelhttp" without types was imported from "github.com/docker/docker/client"
```

要点:
- ツール/依存解決由来の内部エラーが発生し、全件完走不可。
