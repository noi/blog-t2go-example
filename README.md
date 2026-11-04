# blog-t2go-example

ミラティブのテックブログ記事で紹介した、Perl → Go のテスト移植スキル `t2go` のサンプルリポジトリです。社内版からリポジトリ固有の部分を取り除いたスキルと、それを実際に動かすための小さな題材を置いています。

配布して使ってもらうことは想定していません。記事で説明した設計を、実際のファイルとして読めるようにするためのものです。

## 構成

```
.claude/
  skills/port-test/        スキル本体（1ファイルを移植する）
    references/            言語非依存の原則・基本のレビュー観点・知見ループの手順
    templates/             プロジェクトプロファイルの雛形
  skills/port-test-batch/  複数ファイルを並列で移植するコーディネータ
  agents/                  planner / implementer / reviewer
  port-test/               このリポジトリ用のプロファイル（対応表・落とし穴集など）
internal/todo/             題材の TODO API（Go に移植済み）
t/feature/                 移植前の Perl E2E テスト
tests/                     移植後の Go E2E テスト
script/prove.sh            Perl のテストを API サーバーに対して実行する
```

## 題材

`internal/todo` の TODO API は Go に移植済みで、それを検証する E2E テストだけが Perl（`t/feature/todo.t`）のまま残っている、という状況を再現しています。API は旧実装の流儀を引き継ぎ、エラーでも HTTP 200 で本文の `status.ok` を 0 にして返します。

`t/feature/todo.t` には、本番の移植で実際に遭遇したパターンを再現して仕込んでいます。

- `ignore` の後のカンマ抜けで、後続のフィールドが検証されていない subtest
- 削除済みのエンドポイントが 404 を返すことだけを確かめている subtest

## 動かし方

```bash
cpanm -L local --installdeps .     # Perl の依存（Test::Deep）
./script/prove.sh                   # 移植前の Perl テスト
go test ./...                       # Go のテスト
go tool golangci-lint run ./...     # lint
```

このリポジトリで Claude Code を起動し、スキルを実行します。

```
/port-test t/feature/todo.t
```

## ライセンス

[The Unlicense](UNLICENSE)
