# port-test プロファイル

`port-test` / `port-test-batch` スキル用のプロジェクト固有の設定。人が編集する。隣の知見ファイルは自動で育つ。

## 移植

- 言語ペア: Perl → Go
- legacy_test_glob: `t/feature/**/*.t`
- new_test_root: `tests/`
- 読むべきプロジェクトのテスト規約: なし

## 旧実装か移植済みか

このリポジトリの API はすべて Go に移植済み（`internal/todo`）。Perl 側に残っているのはテスト（`t/feature`）だけなので、旧スタックへのブリッジは無い。

- エンドポイント: `internal/todo/handler.go` の `NewHandler` に登録されていれば移植済み。登録されていなければ存在しない（削除済み）
- 呼ぶヘルパ: `testutil.NewServer(t)` でサーバーを立て、`s.CreateUser()` で得たクライアントの `Post` / `Get` を使う

## コマンド

- 新テストのパッケージを1つ実行する: `test-execution.md` を参照
- 旧テストを実行する: `./script/prove.sh t/feature/<name>.t`
- lint: `go tool golangci-lint run ./...`（0 件が完了条件）
- ベースブランチ: `main`

## Git / PR の規約

- 単体実行のブランチ: `port/<短縮名>`
- バッチのブランチ接頭辞: `port/`
- コミットメッセージ / PR タイトル: `port: <旧テストのパス>`
- PR ラベル: `port`        知見 PR のラベル: `port-knowledge`
- 知見ブランチ名のパターン: `*port-knowledge*`
- 新しい worktree にコピーする未追跡ファイル: なし（Perl の依存は `cpanm -L local --installdeps .` で入れる）
- push: サンプルなので push と PR 作成はせず、コミットまでで止める

## バッチの知見

（まだ無い）
