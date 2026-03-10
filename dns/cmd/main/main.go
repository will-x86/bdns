package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"log"

	"github.com/will-x86/bdns/dns/pkg/rcache"
	"github.com/will-x86/bdns/dns/pkg/server"
)

func main() {
	log.Printf("Init valkey")
	err := rcache.InitClient()
	if err != nil {
		log.Fatalf("Failed to initialize valkey client: %v\n", err)
	}
	log.Printf("Starting server...\n")
	server.RunServer()

}
