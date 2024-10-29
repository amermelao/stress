package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v11"
	"log/slog"
	"net/http"
)

type config struct {
	User     string `env:"DB_USER" envDefault:"testuser"`
	Password string `env:"DB_PASSWORD" envDefault:"S3cret"`
	DBName   string `env:"DB_NAME" envDefault:"testcase"`
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
}

var cfg config
var dropDB bool

func init() {
	if err := env.Parse(&cfg); err != nil {
		slog.Error("bad init", "error", err.Error())
	} else {
		slog.Info("config  loaded")
	}

	flag.BoolVar(&dropDB, "drop", false, "drop tables and rebuild")

	flag.Parse()
}

func main() {
	if dropDB {
		dropDatabase()
	} else {
		runApi()
	}
}

func dropDatabase() {
	if db, err := NewConnection(DBSecrets{
		User:     cfg.User,
		Password: cfg.Password,
		DBName:   cfg.DBName,
		Host:     cfg.Host,
		Port:     cfg.Port,
	}); err != nil {
		slog.Error("fail to connect to db",
			"error", err.Error(),
		)
	} else if err := DropTables(db); err != nil {
		slog.Error("fail to drop",
			"error", err.Error(),
		)
	} else if err := AddTables(db); err != nil {
		slog.Error("fail to make",
			"error", err.Error(),
		)
	} else {
		slog.Info("all tables dropped")
	}
}

func runApi() {

	if db, err := NewConnection(DBSecrets{
		User:     cfg.User,
		Password: cfg.Password,
		DBName:   cfg.DBName,
		Host:     cfg.Host,
		Port:     cfg.Port,
	}); err != nil {
		slog.Error("fail to connect to db",
			"error", err.Error(),
		)
	} else {

		stressTests := []GeneralTestCase{
			NewTestCase[NoIndex]("noindex", db),
			NewTestCase[CreateAtUser]("createatuser", db),
			NewTestCase[TSV]("tsv", db),
		}

		base := http.NewServeMux()
		for _, tHandler := range stressTests {
			tHandler.Register(base)
		}

		addr := ":9090"
		slog.Info(fmt.Sprintf("addr: %s", addr))
		slog.Error("fail",
			"error", http.ListenAndServe(addr, logRequests(secure(base))),
		)
	}
}
