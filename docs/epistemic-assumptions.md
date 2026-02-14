# Docksphinx MVP: 認識論的前提と反証条件

## 目的
実装判断に含まれる前提を「検証可能な主張」と「経験則（ヒューリスティック）」に分離し、どの条件で結論が覆るかを明示する。

---

## 1. 検証可能な主張（Fact claims）

1. リポジトリには `cmd/`, `proto/`, `api/` が存在しない。
2. `go test ./...` は `docksphinx/api/docksphinx/v1` 不在で失敗する。
3. `internal/docker` は containers/images/networks/volumes の取得関数を持つ。
4. `internal/monitor` は連続N回判定を実装済みで cooldown は未実装。
5. `internal/grpc` は reflection が常時有効化されている。

---

## 2. ヒューリスティック主張（Heuristic claims）

1. TUIライブラリとして Bubble Tea 系が最小実装リスク。
2. volume usage はMVPで metadata-based 代替が実務上妥当。
3. compose依存可視化はラベル優先の推定がコスト対効果に優れる。
4. stream backpressure は drop-new + dropped_count が運用観測しやすい。

---

## 3. 結論の反証条件（Falsifiability）

以下のいずれかが成立した場合、現行提案の再設計が必要:

1. proto/cmd/api を実装しても `go test ./...` が構造的に通らない。
2. Bubble Tea 実装で pause/reconnect 時に再現性あるフリーズが解消できない。
3. metadata-based volume 指標が運用要件（原因調査初動）を満たさないと実測で判定された。
4. compose推定の誤判定率が高く、調査速度を実際に低下させると確認された。
5. drop-new方針で重要イベント欠落が許容不能と判断された。

---

## 4. 再評価トリガー
- Goツールチェーン更新で `govulncheck` / `gosec` の結果が変化したとき
- Docker API互換変更でメトリクス定義に差分が出たとき
- 要件更新で「volumeの厳密使用量」がMVP必須化されたとき

---

## 5. 参照
- `docs/requirements.md`
- `docs/mvp-spec-freeze.md`
- `docs/mvp-acceptance-matrix.md`
- `docs/security-check-summary.md`
