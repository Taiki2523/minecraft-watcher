package main

import (
	"log"

	"github.com/taiki2523/minecraft-watcher/pkg/chatapp/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("chat server error: %v", err)
	}
}
