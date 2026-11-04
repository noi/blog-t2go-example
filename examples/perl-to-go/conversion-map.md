# 変換対応表: Perl（Test::More / Test::Deep）→ Go（testing / testify / gjson）

プロジェクトの `conversion-map.md` の初期値。扱うのは言語レベル・ライブラリレベルの対応だけ。プロジェクトのヘルパ（テストサーバー・factory・旧スタックへのブリッジ）は各プロジェクトで足す。

## テストの構造

| Perl | Go | 備考 |
|---|---|---|
| `subtest 'name' => sub { ... }` | `t.Run("name", func(t *testing.T) { ... })` | 元の subtest 名を維持する。typo があっても維持する（`misspell` はその行だけ抑制する） |
| `done_testing` | （不要） | |
| `local $ENV{...}` / `mock_guard` のスコープ | セットアップ + `t.Cleanup` | mock のスコープは観測の窓になる。「mock」の節を参照 |
| subtest ごとの DB リセット | `t.Run` ごとにリセットされるとは**限らない** | 新基盤がフィクスチャをテスト単位で1回だけ読むなら、全 subtest の行が共存する。データかフィクスチャのディレクトリを分ける |

## アサーション

| Perl | Go | 備考 |
|---|---|---|
| `ok $x` | `require.True` / `assert.True` | 値の比較を明示するほうがよい。`ok` は弱い |
| `is $got, $exp` | `assert.Equal` / `require.Equal` | `is` は文字列として比較する。Go 側の型は実測で決める（落とし穴「型」） |
| `is $x, undef` | `Type == gjson.Null` **かつ** `.Exists()` | gjson はキーが無いとゼロ値の `Result` を返し、その `Type` は `Null` |
| `cmp_deeply $got, $hash`（plain hashref） | 全キーの完全一致 + キー数の固定 | plain hashref は**キー集合**を完全一致で比較する |
| `superhashof(...)` | パス指定の部分一致 | 列挙したキーだけを見るが、**各キーは存在しなければならない** |
| `bag(...)` | id で引く順不同マッチ + `require.Len` | 順不同だが件数は厳密。index パス（`x.0`・`x.1`）で書くと黙って順序を固定してしまう |
| `subbagof(...)` / `subsetof(...)` | **実測した**集合との完全一致 | actual ⊆ expected の意味なので、**空の結果でも通る**。`Contains` に訳さない |
| `superbagof(...)` | `Subset` + 理由 | 向きは `subbagof` の逆 |
| `ignore()` | `NotZero` / `Exists()` | 値は見ないが、**キーの存在は要求している**。フィールドごと落とすとテストが弱くなる |
| `like $x, qr/.../` | `assert.Regexp` か完全一致 | 正規表現が分岐を区別できているか確認する。共通の接頭辞だけでは何も証明しない |
| `dies_ok` / `throws_ok` | `require.Error` + エラーの型/メッセージ | |

## レスポンス（JSON）

- Perl の API はエラーでも **HTTP 200 で本文中にステータス**（`status.ok = 0`）を返すことが多い。HTTP コードだけでなく本文のステータスも見る。
- 他にレスポンスを検証しない呼び出しでは、成功フラグを **fatal で**アサートする: `require.EqualValues(t, 1, gjson.Get(res, "status.ok").Int())`。
- gjson:
  - キーが無いとゼロ値の `Result` を返す。`.Int()` / `.String()` / `.Bool()` はゼロ値になる。「無い」ことがバグになる箇所では `.Exists()` を併記する。
  - `path.#` は `path` が無いと 0 を返す。0 件のアサーションには成功ガードが要る。
  - `path.#(cond)` は**最初の1件**しか返さない。マッチ数を数えるなら `#(cond)#` を使う。
  - `gjson.Result` は 80 バイトで、gocritic の `hugeParam` は 80 以上で発火する。ヘルパにはポインタで渡す。
- レスポンスから取り出した ID は、使う前に `require.NotEmpty` でガードする。パースに失敗すると 0 が返り、多くの factory は 0 を「新しく作る」と解釈する。

## mock → 実際の作用

| Perl のパターン | Go |
|---|---|
| enqueue を mock し、累積配列全体に `cmp_deeply` | **実**キューで drain → 操作 → drain。窓の総件数を `require.Len` で固定し、**かつ**セットアップ中に該当種別が 0 件だったこともアサートする |
| push / logger の mock | テスト環境に実際の受け先があればそれを読む |
| `dequeue(sub { is undef, shift })` | 注意: キューが**空だとコールバックを呼ばない**ヘルパがある。その場合、直後の `cmp_deeply $row` は前回の行を再比較している。drain して数える形に置き換える |

## 時刻

| Perl | Go | 備考 |
|---|---|---|
| `set_fixed_time($t)` | プロジェクトのモック時刻 | 常に絶対オフセット |
| ヘルパ内の `my $time = time;` | — | 外側の `$time` を**シャドーイング**するだけで進めない。Go のヘルパに「新しい時刻」を返させて次のフェーズの起点にしない |
| `Sub::Retry::retry($times, $delay, \&code, \&retry_if)` | テスト goroutine 内の deadline ループ | 3番目が実行対象、4番目が継続判定。条件が `t.Fatal` しうるなら `require.Eventually` は使わない（落とし穴を参照） |

## エンコーディング

- Perl のテストファイルには普通 `use utf8;` がある。Go のブリッジが実行する生成 Perl スニペットには**無いことが多い**。無いと、非 ASCII のリテラルはバイト列になる:
  - DB に書いた値が二重エンコードされる。
  - デコード済みのテキストに対する正規表現が決してマッチしない。**否定のマッチは常に通る。**
- `binmode STDOUT, ':utf8'` なしで非 ASCII を `print` すると「Wide character」警告が出て、UTF-8 の出力が保証されない。
- JSON のレスポンスは `\uXXXX` エスケープを含むことがあり、生の本文に非 ASCII のテキストを当てても失敗する。先に JSON をデコードする。

## 生成 Perl に埋め込む文字列

値は**シングルクォート**にする。ダブルクォートの中では `@x` と `$x` が展開され、未定義の変数は黙って空文字になる。ダブルクォートや `join` は `\n` / `\r\n` が必要な箇所だけで使う。

## 決定性

- Perl のハッシュ順序はプロセスごとにランダム化される。factory がハッシュを反復して入れた行に対し、`ORDER BY` なしの `SELECT ... LIMIT 1` をするリポジトリのコードは非決定になる。行を一意にするか、同じ内容に揃える。
- Go の `map` の反復順もランダム。順序を前提にしない。
