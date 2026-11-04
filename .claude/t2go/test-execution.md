# テスト実行

## パッケージを1つ実行する

```bash
go test ./tests/<name>/... -count=1
```

サーバーは `testutil.NewServer(t)` がテストごとに `httptest` で立てるので、事前の起動は要らない。

## 旧テスト（Perl）を実行する

```bash
cpanm -L local --installdeps .   # 初回のみ
./script/prove.sh t/feature/<name>.t
```

`script/prove.sh` が API サーバーをビルドして起動し、`TEST_BASE_URL` を渡して `prove` を実行する。移植前に旧テストの実際の挙動を確かめるときに使う。

## 長時間の実行

テストは数秒で終わるので不要。

## 並列実行（バッチ）

- 同時実行どうしの分離: `NewServer` がテストごとに独立したサーバーとメモリ上の状態を持つので干渉しない
- 旧テストの `script/prove.sh` は固定ポート（18080）を使うので、同時に2本走らせない

## 事前チェック（バッチ B0）

- 特になし

## トラブルシューティング

- 手元の golangci-lint が古いと Go 1.27 の export data を読めずに typecheck エラーになる。必ず `go tool golangci-lint`（go.mod で固定した版）を使う
