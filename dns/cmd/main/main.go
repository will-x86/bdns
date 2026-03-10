package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"context"
	"log"

	"github.com/will-x86/bdns/dns/pkg/server"
)

func main() {
	log.Printf("Starting server...\n")
	ctx := context.Background()
	server.RunServer(ctx, "server.crt", "server.key")

}
