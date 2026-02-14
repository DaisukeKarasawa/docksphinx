# セキュリティベースラインチェックリスト（MVP）

## 目的
ローカル常駐前提のMVPで最低限守るべきセキュリティ基準を運用で確認する。

---

## 1. 通信面
- [ ] gRPC bind が `127.0.0.1` 既定
- [ ] 外部公開ポート設定が無効
- [ ] gRPC reflection がデフォルト無効

---

## 2. 設定/ファイル権限
- [ ] config ファイル権限が適切（機密情報を含む場合は最小権限）
- [ ] PIDファイルの配置先/権限が適切
- [ ] ログファイルに過剰情報が出ていない

---

## 3. 入力バリデーション
- [ ] 正規表現フィルタのcompileチェックがある
- [ ] interval/threshold/max_history の範囲チェックがある
- [ ] path引数の異常値が安全に失敗する

---

## 4. 安定性（DoS的自己劣化防止）
- [ ] stream購読解除（unsubscribe）が徹底されている
- [ ] subscriber遅延時のdrop方針がある
- [ ] goroutine/chan close順序が設計されている

---

## 5. エラー情報露出
- [ ] panic/stacktraceを通常運用ログへ出さない
- [ ] 環境変数や絶対パスを過剰に露出しない

---

## 6. 実施ログ
- [ ] 実行コマンドを `docs/validation-log-template.md` に記録
- [ ] 指摘事項を `docs/risk-register.md` に反映
