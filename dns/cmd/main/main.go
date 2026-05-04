package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/UnnoTed/horizontal"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	"codeberg.org/will-x86/bdns/dns/pkg/server"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
)

func main() {

	// flags
	ingest := flag.Bool("ingest", false, "download and ingest StevenBlack blocklist")
	seed := flag.Bool("seed", false, "seed with init.sql")
	flag.Parse()
	//config
	config, log := configAndLogger()

	// db setup + ingest
	if err := db.InitDB(log, "./app.db", "./migrations/"); err != nil {
		log.Fatal().Err(err).Msg("failed to initalize datavase")
	}
	if *seed {
		db.Seed(log)
	}

	if *ingest {
		log.Debug().Msg("ingesting blocklist")
		if err := db.Ingest(); err != nil {
			log.Fatal().Err(err).Msg("ingest failed")
		}
		log.Debug().Msg("ingesting done")
	}

	ctx := context.Background()
	ctx = log.WithContext(ctx)
	server.RunServer(ctx, &config)
}
func configAndLogger() (server.ServerConfig, zerolog.Logger) {
	c := server.ServerConfig{
		PrivateKey: os.Getenv("KEY_PATH"),
		SignedKey:  os.Getenv("CRT_PATH"),
		ValkeyAddr: os.Getenv("VALKEY_ADDR"),
	}
	c.Port = func() int {
		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			return 8533
		}
		return port

	}()
	// default to warn
	logLevel := func() zerolog.Level {
		level := os.Getenv("LOG_LEVEL")
		level = strings.ToLower(level)
		switch level {
		case "panic":
			return zerolog.PanicLevel
		case "fatal":
			return zerolog.FatalLevel
		case "error":
			return zerolog.ErrorLevel
		case "warn":
			return zerolog.WarnLevel
		case "info":
			return zerolog.InfoLevel
		case "debug":
			return zerolog.DebugLevel
		case "trace":
			return zerolog.TraceLevel
		default:
			return zerolog.WarnLevel
		}
	}()
	var log zerolog.Logger
	/*log.Logger = log.Output(horizontal.ConsoleWriter{Out: os.Stderr})
	log.Debug().Msg("hi")
	log.Debug().Msg("hello")*/
	zerolog.SetGlobalLevel(logLevel)
	{
		eKey := "ENVIRONMENT" // I really can't trust myself to spell
		// Hopefully this makes it a little obvious we're using "production" as a key
		if os.Getenv(eKey) == "" || os.Getenv(eKey) == "production" {
			os.Setenv(eKey, "production")
			log = zerolog.New(os.Stdout).With().Timestamp().Logger()
		} else if os.Getenv(eKey) == "local" {
			// local, pretty !!
			log = log.Output(horizontal.ConsoleWriter{Out: os.Stdout}).Level(logLevel)
		}
	}
	log.Info().Any("config", c).Msg("Starting server")
	log.Debug().Str("log_level", logLevel.String()).Msg("log level:")
	log.Trace().Msg("tracing is enabled")
	return c, log
}
