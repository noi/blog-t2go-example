package me_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"blog-t2go-example/tests/testutil"
)

func TestMe(t *testing.T) {
	t.Parallel()

	t.Run("me", func(t *testing.T) {
		t.Parallel()
		s := testutil.NewServer(t)
		user := s.CreateUser()

		code, res := user.Get(t, "/api/me")

		require.Equal(t, http.StatusOK, code)
		assert.Equal(t, user.UserID(), gjson.Get(res, "user.id").String())
	})
}
