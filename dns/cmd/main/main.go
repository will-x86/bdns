package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"context"
	"flag"
	"log"

	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/server"
)

func main() {
	ingest := flag.Bool("ingest", false, "download and ingest StevenBlack social blocklist")
	flag.Parse()

	if err := db.InitDB("./app.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}

	if *ingest {
		log.Println("Ingesting StevenBlack social blocklist...")
		if err := db.IngestStevenBlack(); err != nil {
			log.Fatalf("Ingest failed: %v\n", err)
		}
		log.Println("Done.")
		return
	}

	s, e := db.CreateUser()
	log.Printf("Created user: %s, error: %v\n", s, e)
	log.Printf("Starting server...\n")
	ctx := context.Background()

	server.RunServer(ctx, "server.crt", "server.key")
}
