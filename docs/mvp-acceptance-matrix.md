# Docksphinx MVP 受け入れ基準マトリクス

## 目的
`docs/requirements.md` のMVP成功条件と、実装提案（outputs）の対応関係を明示し、レビュー時の抜け漏れを防ぐ。

---

## A) docksphinxd（デーモン）
- 受け入れ基準:
  - 起動/停止/状態確認
  - graceful shutdown
- 対応成果:
  - `outputs/phase1-implementation.md`
  - `outputs/patches/phase1-2.diff.md`

## B) 設定ロード/永続化
- 受け入れ基準:
  - YAMLロード
  - デフォルト
  - バリデーション
  - パス解決
- 対応成果:
  - `outputs/phase0-design-memo.md`
  - `outputs/phase1-implementation.md`
  - `outputs/patches/phase1-2.diff.md`

## C) ログとイベント履歴
- 受け入れ基準:
  - 稼働ログ
  - max_historyメモリ保持
- 対応成果:
  - `outputs/phase2-implementation.md`
  - `outputs/patches/phase1-2.diff.md`

## D) CLI健全化（snapshot/tail）
- 受け入れ基準:
  - context/defer/close
  - 再接続/購読解除
  - 長時間劣化防止
- 対応成果:
  - `outputs/phase2-implementation.md`
  - `outputs/patches/phase1-2.diff.md`
  - `outputs/phase5-validation.md`

## E) TUI（4ペイン+操作）
- 受け入れ基準:
  - 4ペイン
  - Tab/矢印/jk/検索/ソート/pause/q
  - stream更新/詳細/イベント
- 対応成果:
  - `outputs/phase3-implementation.md`
  - `outputs/specs/tui-interaction.md`
  - `outputs/patches/phase3-6.diff.md`

## F) image/network/volume収集
- 受け入れ基準:
  - 収集頻度/保持/表示
  - 左ペイン切替に反映
- 対応成果:
  - `outputs/phase4-implementation.md`
  - `outputs/specs/metrics-definition.md`
  - `outputs/patches/phase3-6.diff.md`

## G) uptime / volume usage
- 受け入れ基準:
  - uptime表示
  - volume代替指標の仕様固定
- 対応成果:
  - `docs/requirements.md`（MVP固定判断追記）
  - `outputs/specs/metrics-definition.md`
  - `outputs/patches/phase3-6.diff.md`

---

## H1) launchd運用
- 対応成果:
  - `outputs/ops/launchd.md`
  - `outputs/phase2-implementation.md`

## H2) リーク/バックプレッシャ
- 対応成果:
  - `outputs/phase0-design-memo.md`
  - `outputs/patches/phase1-2.diff.md`
  - `outputs/patches/phase3-6.diff.md`

## H3) 閾値ノイズ対策（cooldown）
- 対応成果:
  - `docs/requirements.md`（運用最適化追記）
  - `outputs/patches/phase3-6.diff.md`

## H4) Composeまとまり可視化
- 対応成果:
  - `docs/requirements.md`（MVP方針追記）
  - `outputs/phase4-implementation.md`
  - `outputs/patches/phase3-6.diff.md`

## H5) メトリクス定義固定
- 対応成果:
  - `docs/requirements.md`
  - `outputs/specs/metrics-definition.md`

## H6) テスト戦略
- 対応成果:
  - `outputs/phase5-validation.md`
  - `outputs/test-plan.md`

---

## Security / Vulnerability Gate
- 対応成果:
  - `outputs/security-report.md`
  - `outputs/phase6-security-refactor.md`

## Refactoring Gate
- 対応成果:
  - `outputs/refactor-report.md`
  - `outputs/phase6-security-refactor.md`

---

## 補足
- 本マトリクスは「直接コード変更禁止」の運用ルール下で、実装適用可能性を担保するための追跡表である。
