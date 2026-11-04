// Package todo は TODO API の実装。Perl から Go に移植済みで、レスポンスは旧実装の
// 流儀を引き継ぎ、エラーでも HTTP 200 + 本文の status.ok=0 で返す。
package todo

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Todo は1件の TODO。
type Todo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt int64  `json:"created_at"`
}

type store struct {
	mu     sync.Mutex
	nextID int64
	todos  map[string][]*Todo
}

// NewHandler は API のハンドラを返す。状態はプロセス内のメモリに持つ。
func NewHandler() http.Handler {
	s := &store{todos: map[string][]*Todo{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/me", s.me)
	mux.HandleFunc("POST /api/todo/create", s.create)
	mux.HandleFunc("GET /api/todo/list", s.list)
	mux.HandleFunc("POST /api/todo/done", s.done)
	return mux
}

func (s *store) me(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeError(w, "login required")
		return
	}
	writeOK(w, map[string]any{"user": map[string]any{"id": userID}})
}

func (s *store) create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeError(w, "login required")
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request")
		return
	}
	if req.Title == "" {
		writeError(w, "title is required")
		return
	}

	s.mu.Lock()
	s.nextID++
	t := &Todo{ID: s.nextID, Title: req.Title, CreatedAt: time.Now().Unix()}
	s.todos[userID] = append(s.todos[userID], t)
	s.mu.Unlock()

	writeOK(w, map[string]any{"todo": t})
}

func (s *store) list(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeError(w, "login required")
		return
	}
	s.mu.Lock()
	todos := append([]*Todo{}, s.todos[userID]...)
	s.mu.Unlock()

	writeOK(w, map[string]any{"todos": todos})
}

func (s *store) done(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		writeError(w, "login required")
		return
	}
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.todos[userID] {
		if t.ID == req.ID {
			t.Done = true
			writeOK(w, map[string]any{"todo": t})
			return
		}
	}
	writeError(w, "todo not found")
}

func writeOK(w http.ResponseWriter, body map[string]any) {
	body["status"] = map[string]any{"ok": 1}
	writeJSON(w, body)
}

func writeError(w http.ResponseWriter, msg string) {
	writeJSON(w, map[string]any{"status": map[string]any{"ok": 0, "error": msg}})
}

func writeJSON(w http.ResponseWriter, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}
