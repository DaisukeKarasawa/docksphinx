# ADR-0001: MVP Delivery Mode under Repository Policy

- Status: Accepted
- Date: 2026-02-14

## Context

本リポジトリの運用ルール（AGENTS.md）では、コード直接変更と実装の直接投入が禁止されている。  
一方、MVP要求は `docs/requirements.md` に基づく具体実装を必要とする。

この制約下で、成果の再現性・検証可能性・レビュー可能性を落とさずに進める必要がある。

## Decision

MVPの実装成果は以下の二層で提供する。

1. `outputs/`（非追跡）:
   - フェーズ別実装手順書
   - 具体差分（diff）提案
   - 検証コマンドと実行結果
2. `docs/`（追跡）:
   - 要件固定化
   - 受け入れ基準マトリクス
   - 適用ランブック
   - セキュリティ/検証サマリ

## Consequences

### Positive
- 運用ルール違反なく、実装可能な成果を継続提供できる。
- レビュー導線（README→docs→outputs）が確立される。
- 実装担当者へのハンドオフ品質が向上する。

### Negative
- 実コード未適用のため、最終的なコンパイル・実行確認は別工程になる。
- `outputs/` がgit非追跡のため、成果共有には追加手順が必要。

## Alternatives Considered

1. 直接コード変更:
   - 却下（運用ルール違反）
2. docsのみ更新（diff提案なし）:
   - 却下（実装再現性が不足）

## Verification

本ADRの有効性は次で確認する:
- 実装担当が `docs/patch-application-procedure.md` の手順で差分適用できる
- `docs/mvp-acceptance-matrix.md` に沿って要件照合できる
