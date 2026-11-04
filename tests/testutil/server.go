// Package testutil は Go の E2E テストの共通ヘルパ。
package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"blog-t2go-example/internal/todo"
)

var userSeq atomic.Int64

// Server はテスト用に起動した API サーバー。
type Server struct {
	t       *testing.T
	baseURL string
}

// NewServer はテストごとに独立した API サーバーを起動する。
func NewServer(t *testing.T) *Server {
	t.Helper()
	srv := httptest.NewServer(todo.NewHandler())
	t.Cleanup(srv.Close)
	return &Server{t: t, baseURL: srv.URL}
}

// CreateUser はログイン済みユーザーのクライアントを返す。ユーザー ID はテスト内で一意。
func (s *Server) CreateUser() *Client {
	return &Client{s: s, userID: fmt.Sprintf("user-%d", userSeq.Add(1))}
}

// Client はユーザーとして API を呼ぶ。
type Client struct {
	s      *Server
	userID string
}

// UserID はクライアントのユーザー ID を返す。
func (c *Client) UserID() string { return c.userID }

// Post は JSON を POST し、HTTP 200 を確認してから本文を返す。
func (c *Client) Post(t *testing.T, path string, params any) string {
	t.Helper()
	body, err := json.Marshal(params)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, c.s.baseURL+path, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return c.do(t, req)
}

// Get は GET し、HTTP 200 を確認してから本文を返す。
func (c *Client) Get(t *testing.T, path string) string {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, c.s.baseURL+path, http.NoBody)
	require.NoError(t, err)
	return c.do(t, req)
}

func (c *Client) do(t *testing.T, req *http.Request) string {
	t.Helper()
	req.Header.Set("X-User-Id", c.userID)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, res.Body.Close())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode, "body: %s", body)
	return string(body)
}
