# Docksphinx MVP 用語集

## Docksphinx
Docker環境の状態監視を行うCLI/TUIツール本体。

## docksphinxd
常駐監視デーモン。Docker APIから収集し、gRPCで配信する。

## snapshot
現在状態を1回だけ取得するAPI/CLI操作。

## tail
イベントを継続購読するCLI表示モード。

## TUI
ターミナルUI。4ペインで状態を可視化する。

## Backpressure
配信先が遅いときに発生する詰まり圧力。  
本MVP提案では非ブロッキング配信+drop方針で吸収する。

## Cooldown
同一イベントの連投を抑えるための最低間隔。

## Compose grouping
`com.docker.compose.project/service` ラベル等から推定するまとまり表示。

## Metadata-based volume indicator
volumeの厳密使用量ではなく、mount件数やname/source/destination/driverなどで代替表示するMVP方針。

## outputs成果
運用ルール上、直接コード変更の代わりに生成する実装手順書・diff提案・検証報告群。
