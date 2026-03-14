package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/server"
)

func getConfig() server.ServerConfig {
	c := server.ServerConfig{
		PrivateKey: os.Getenv("KEY_PATH"),
		SignedKey:  os.Getenv("CRT_PATH"),
	}

	c.ValkeyAddr = func() string {
		vA := os.Getenv("VALKEY_ADDR")
		if vA == "" {
			return "localhost:6379"
		}
		return vA
	}()
	c.Port = func() int {
		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			return 8533
		}
		return port

	}()
	return c
}
func main() {
	ingest := flag.Bool("ingest", false, "download and ingest StevenBlack blocklist")
	seed := flag.Bool("seed", false, "seed with init.sql")
	flag.Parse()
	config := getConfig()
	if err := db.InitDB("./app.db", "./migrations/"); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	if *seed {
		db.Seed()
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
