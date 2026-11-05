package todo_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"blog-t2go-example/tests/testutil"
)

// 旧テストの subtest 'archive' は移植しない。NewHandler に /api/todo/archive は登録されておらず、
// 旧テストは存在しないルートが 404 になることしか見ていなかった。

// createTodo は TODO を作成し、作成された todo の JSON を返す。
func createTodo(t *testing.T, user *testutil.Client, title string) gjson.Result {
	t.Helper()
	code, res := user.Post(t, "/api/todo/create", map[string]any{"title": title})
	require.Equal(t, http.StatusOK, code)
	return gjson.Get(res, "todo")
}

// keys はオブジェクトのキーをソートして返す。
func keys(r gjson.Result) []string {
	var ks []string
	r.ForEach(func(k, _ gjson.Result) bool {
		ks = append(ks, k.String())
		return true
	})
	slices.Sort(ks)
	return ks
}

func TestTodo(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		s := testutil.NewServer(t)
		user := s.CreateUser()

		code, res := user.Post(t, "/api/todo/create", map[string]any{"title": "牛乳を買う"})

		require.Equal(t, http.StatusOK, code)
		assert.Equal(t, []string{"todo"}, keys(gjson.Parse(res)))
		todo := gjson.Get(res, "todo")
		assert.Equal(t, []string{"created_at", "done", "id", "title"}, keys(todo))
		assert.Equal(t, gjson.Number, todo.Get("id").Type)
		assert.NotZero(t, todo.Get("id").Int())
		assert.Equal(t, "牛乳を買う", todo.Get("title").String())
		assert.Equal(t, gjson.False, todo.Get("done").Type)
		assert.Equal(t, gjson.Number, todo.Get("created_at").Type)
		assert.NotZero(t, todo.Get("created_at").Int())
	})

	t.Run("create without title", func(t *testing.T) {
		t.Parallel()
		s := testutil.NewServer(t)
		user := s.CreateUser()

		code, res := user.Post(t, "/api/todo/create", map[string]any{"title": ""})

		require.Equal(t, http.StatusBadRequest, code)
		assert.Equal(t, "title is required", gjson.Get(res, "error").String())
	})

	t.Run("list", func(t *testing.T) {
		t.Parallel()
		s := testutil.NewServer(t)
		user := s.CreateUser()
		for _, title := range []string{"a", "b", "c"} {
			createTodo(t, user, title)
		}
		other := s.CreateUser()
		createTodo(t, other, "other")

		code, res := user.Get(t, "/api/todo/list")

		require.Equal(t, http.StatusOK, code)
		todos := gjson.Get(res, "todos").Array()
		require.Len(t, todos, 3)
		var titles []string
		for _, todo := range todos {
			titles = append(titles, todo.Get("title").String())
		}
		assert.ElementsMatch(t, []string{"a", "b", "c"}, titles)

		code, res = other.Get(t, "/api/todo/list")

		require.Equal(t, http.StatusOK, code)
		otherTodos := gjson.Get(res, "todos").Array()
		require.Len(t, otherTodos, 1)
		assert.Equal(t, "other", otherTodos[0].Get("title").String())
	})

	t.Run("done", func(t *testing.T) {
		t.Parallel()
		s := testutil.NewServer(t)
		user := s.CreateUser()
		id := createTodo(t, user, "done me").Get("id").Int()
		require.NotZero(t, id)

		code, res := user.Post(t, "/api/todo/done", map[string]any{"id": id})

		require.Equal(t, http.StatusOK, code)
		todo := gjson.Get(res, "todo")
		assert.Equal(t, id, todo.Get("id").Int())
		assert.Equal(t, "done me", todo.Get("title").String())
		// 旧テストは ignore の後のカンマ抜けで done が検証されていなかったため、ここで明示する。
		assert.Equal(t, gjson.True, todo.Get("done").Type)
	})

	t.Run("done by other user", func(t *testing.T) {
		t.Parallel()
		s := testutil.NewServer(t)
		user := s.CreateUser()
		id := createTodo(t, user, "mine").Get("id").Int()
		require.NotZero(t, id)
		other := s.CreateUser()

		code, res := other.Post(t, "/api/todo/done", map[string]any{"id": id})

		require.Equal(t, http.StatusNotFound, code)
		assert.Equal(t, "todo not found", gjson.Get(res, "error").String())

		code, res = user.Get(t, "/api/todo/list")
		require.Equal(t, http.StatusOK, code)
		todos := gjson.Get(res, "todos").Array()
		require.Len(t, todos, 1)
		assert.Equal(t, id, todos[0].Get("id").Int())
		assert.Equal(t, gjson.False, todos[0].Get("done").Type)

		code, _ = user.Post(t, "/api/todo/done", map[string]any{"id": id})
		require.Equal(t, http.StatusOK, code)
	})
}
