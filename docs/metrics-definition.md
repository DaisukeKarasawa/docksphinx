# メトリクス定義（MVP）

## 1. 目的
MVPで表示する CPU / Memory / Network / Uptime / Volume の定義と欠損時の表示規則を固定し、  
OS差異や Docker API 制約による曖昧さをなくします。

## 2. 取得周期
- `monitor.interval`:
  - コンテナ一覧、状態、CPU/Memory/Network を取得
- `monitor.resource_interval`:
  - images / networks / volumes / compose grouping を取得

## 3. 各メトリクス定義

### 3.1 CPU%
- 取得元: `ContainerStats` (`cpuDelta/systemDelta * numCPU * 100`)
- 単位: %
- 用途: ソート、閾値判定（`cpu_threshold`）
- 欠損時: `0` ではなく `N/A` 相当で扱う（TUIでは未取得時に前回値が無い場合 0 表示になり得る）

### 3.2 Memory
- `memory_usage`: bytes
- `memory_limit`: bytes
- `memory_percent`: `usage/limit * 100`
- 用途: ソート、閾値判定（`mem_threshold`）
- 欠損時: `N/A`

### 3.3 Network Rx/Tx
- 取得元: `ContainerStats.Networks` の全インターフェース合計
- `network_rx`: 受信 bytes
- `network_tx`: 送信 bytes
- 欠損時: `N/A`

### 3.4 Uptime
- 取得元: `ContainerInspect.State.StartedAt`
- `uptime_seconds`: `now - started_at`
- 停止中コンテナでは `StartedAt` 不在または古い値になり得るため、状態と合わせて解釈
- 欠損時: `N/A`

### 3.5 Volume usage（MVP制約）
- Docker Engine API だけで「正確な容量使用量（bytes）」を安定取得することは困難。
- MVPでは以下を代替採用:
  - volume name / driver / mountpoint
  - ref_count（取得可能な場合）
  - `usage_note = metadata-only (size unavailable via Docker API)`
- したがって、**MVPの volume は容量メータではなくメタデータ監視**として扱う。

## 4. Compose grouping 推定
- 第1優先: ラベル
  - `com.docker.compose.project`
  - `com.docker.compose.service`
- フォールバック: ラベル未設定時、非システムネットワーク名でヒューリスティック grouping
- 制約:
  - 厳密な依存解析ではない
  - 誤検知・過剰集約の可能性あり

## 5. OS差異と制約
- CPU / Network の値は Docker Desktop / Engine 実装差で差異が出る可能性がある。
- 本MVPは「ローカル監視の初動調査」を目的とし、厳密会計値ではなく傾向把握を優先する。
