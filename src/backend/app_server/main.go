package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/app_server/websocket"
)

func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setupRoutes() {
	http.HandleFunc("/helloWorld", func(w http.ResponseWriter, r *http.Request) {
		list := []string{r.RequestURI}
		a, _ := json.Marshal(list)
		w.Write(a)
	})
}

func main() {
	fmt.Println("Distributed Chat App v0.01")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
