package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"context"
	"log"

	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/server"
)

func main() {
	if err := db.InitDB("./app.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	s, e := db.CreateUser()
	log.Printf("Created user: %s, error: %v\n", s, e)
	log.Printf("Starting server...\n")
	ctx := context.Background()

	server.RunServer(ctx, "server.crt", "server.key")

}
