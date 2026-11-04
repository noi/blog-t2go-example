// Command server は TODO API サーバーを起動する。
package main

import (
	"flag"
	"log"
	"net/http"

	"blog-t2go-example/internal/todo"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	flag.Parse()

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, todo.NewHandler()))
}
