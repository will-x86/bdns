package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	//	_ "github.com/joho/godotenv/autoload"
	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/server"
)

func main() {
	ingest := flag.Bool("ingest", false, "download and ingest StevenBlack blocklist")
	flag.Parse()

	config := server.ServerConfig{}
	if os.Getenv("KEY_PATH") == "" {
		config.PrivateKey = "server.key"
		config.SignedKey = "server.crt"
	} else {
		config.PrivateKey = os.Getenv("KEY_PATH")
		config.SignedKey = os.Getenv("CRT_PATH")
	}
	if os.Getenv("PORT") == "" {
		config.Port = 8533
	} else {
		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			config.Port = 8533
		}
		config.Port = port
	}

	if err := db.InitDB("./app.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}

	if *ingest {
		log.Println("Ingesting blocklist...")
		if err := db.Ingest(); err != nil {
			log.Fatalf("Ingest failed: %v\n", err)
		}
		log.Println("Done.")
		return
	}

	log.Printf("Starting server...\n")
	ctx := context.Background()
	server.RunServer(ctx, &config)
}
