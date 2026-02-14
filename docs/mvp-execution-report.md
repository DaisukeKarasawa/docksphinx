# Docksphinx MVP 実行レポート（運用ルール準拠版）

## 1. 実行方式
本リポジトリでは `AGENTS.md` によりコード直接変更が禁止されているため、MVP実装は次の代替成果で実施した。

- `outputs/` に実装手順書（phase0〜phase6）
- `outputs/` に具体パッチ案（phase1-2 / phase3-6）
- `outputs/` にテスト計画・セキュリティ診断・リファクタ報告

---

## 2. 要件カバレッジ

## MVP A〜G
- A: daemon本体（起動/停止/状態確認/graceful shutdown）提案済み
- B: YAML設定ロード/デフォルト/バリデーション提案済み
- C: ログ/イベント履歴（軽量）提案済み
- D: snapshot/tail健全化提案済み
- E: TUI 4ペイン/操作仕様提案済み
- F: images/networks/volumes 収集と表示提案済み
- G: uptime実装 + volume代替指標方針を仕様化

## Hard Tasks H1〜H6
- H1: launchd運用手順/plist提案済み
- H2: streamバックプレッシャ/購読解除方針提案済み
- H3: 閾値cooldown（連投抑止）提案済み
- H4: composeグルーピング（ラベル優先）提案済み
- H5: メトリクス定義（欠損表示含む）固定済み
- H6: テスト戦略・最小E2E方針提案済み

---

## 3. 実行コマンド結果（実測）
- `go test ./...`:
  - 失敗（`docksphinx/api/docksphinx/v1` 欠落）
- `go test ./internal/docker ./internal/event ./internal/monitor`:
  - 成功
- `go test -race ./internal/docker ./internal/event ./internal/monitor`:
  - 成功
- `govulncheck ./...`:
  - ツール導入後も依存解析内部エラー
- `staticcheck ./...`:
  - ST1005/SA1019/SA4006 + grpc生成物不足を検出
- `gosec ./...`:
  - G115（uint64→int64変換）複数 + grpc解析時panic

---

## 4. 既知制約
- 直接コード変更禁止ルールにより、実コード適用は未実施（適用可能diffとして提供）。
- volume usageはMVPで metadata-based 代替指標を採用。
- 脆弱性スキャンの一部は環境/依存要因で完走不可。再実行条件を報告済み。

---

## 5. 適用順（実装者向け）
1. `outputs/patches/phase1-2.diff.md`
2. `outputs/patches/phase3-6.diff.md`
3. `make proto`
4. `go test ./...`
5. `go test -race ./...`
6. `govulncheck ./... && staticcheck ./... && gosec ./...`
