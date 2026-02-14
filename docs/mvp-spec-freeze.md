# Docksphinx MVP 仕様固定メモ

## 目的
MVP範囲で実装判断を固定し、レビュー時の期待値を一致させる。

---

## 1. スコープ境界
- 対象: `docs/requirements.md` の MVP 範囲のみ
- 非対象:
  - クラウド送信
  - 自動修復（kill/restart等）
  - MCP化
  - GUIアプリ化

---

## 2. 収集対象
- containers
- images
- networks
- volumes

コンテナ収集項目:
- container id/name/image
- cpu%
- memory usage/%
- uptime
- network rx/tx
- volume metadata（MVP代替）

---

## 3. volume usage のMVP判断
- 正確な容量使用量は Docker API のみでは安定取得が困難。
- MVPでは代替として以下を採用:
  - mount count
  - mount metadata（name/source/destination/driver）
- UIに「metadata-only」であることを明示する。

---

## 4. 閾値イベント
- 連続N回で確定
- ノイズ抑制として cooldown を導入（同一イベントの連投抑止）

---

## 5. セキュリティ既定
- gRPC bind は `127.0.0.1` 既定
- gRPC reflection は default off
- 設定値（regex/path/value）はバリデーションする

---

## 6. TUI必須操作
- Tab/矢印
- j/k
- `/`検索
- ソート切替
- pause（表示明示）
- q終了

---

## 7. 常駐運用
- daemon起動/停止/状態確認を提供
- macOSは launchd 運用を手順化
