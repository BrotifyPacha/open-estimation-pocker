package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"estimation-poker/internal/infrastructure/queue"
	"estimation-poker/internal/infrastructure/websockets"

	"golang.org/x/net/websocket"
)

type Mode int

var (
	ModeServer Mode = 0
	ModeClient Mode = 1
)

type cmdArgs struct {
	AppMode Mode
}

var address = "0.0.0.0:8080"

func main() {

	ctx := context.Background()

	appCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	StartHTTPServer(appCtx)
}

func StartHTTPServer(ctx context.Context) {
	queue := queue.NewChanQueue()

	websockets := websockets.NewServer(ctx, queue)

	http.Handle("/{room_id}/ws", websocket.Handler(websockets.WebsocketHandler))

	http.HandleFunc("/", serveFile("./ui/home.html"))
	http.HandleFunc("/{room_id}", serveFile("./ui/room.html"))
	http.HandleFunc("/ui.js", serveFile("./ui/ui.js"))

	server := http.Server{
		Addr:    address,
		Handler: http.DefaultServeMux,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(ctx)
	}()

	log.Println("running server")
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}
