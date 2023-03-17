package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Netflix/go-env"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	web "github.com/sina-am/social-media/common"
	"go.uber.org/zap"
)

type Settings struct {
	PostgresUsername string `env:"POSTGRES_USERNAME,default=postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD,default=1234"`
	PostgresHostname string `env:"POSTGRES_HOSTNAME,default=localhost"`
	PostgresDatabase string `env:"POSTGRES_DATABASE,default=auth"`
	HTTPAddress      string `env:"HTTP_ADDRESS,default=:8000"`
	GRPCAddress      string `env:"GRPC_ADDRESS,default=:5000"`
}

func (s *Settings) GetDatabaseConnStr() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		s.PostgresUsername,
		s.PostgresPassword,
		s.PostgresHostname,
		s.PostgresDatabase,
	)
}

func OverwriteWithSettingFromCli(settings *Settings) {
	address := flag.String("address", ":8000", "listening address for API server")

	flag.Parse()

	if address != nil {
		settings.HTTPAddress = *address
	}
}

func main() {
	var settings Settings
	_, err := env.UnmarshalFromEnviron(&settings)
	if err != nil {
		log.Fatal(err)
	}

	OverwriteWithSettingFromCli(&settings)

	logger, _ := zap.NewProduction()

	storage, err := NewPostgresStorage(settings.GetDatabaseConnStr())
	if err != nil {
		log.Fatal(err)
	}

	reg := prometheus.NewRegistry()
	service := NewMonitorAuthService(NewLocalAuthService(storage, "verysecretkey"), reg)
	apiServer := APIServer{
		APIServer: web.APIServer{
			Addr:   settings.HTTPAddress,
			Logger: logger,
		},
		Service: service,
		Storage: storage,
		Router:  mux.NewRouter(),
	}
	apiServer.Router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	grpcServer := GRPCServer{
		Service: service,
		Addr:    settings.GRPCAddress,
	}
	go func() {
		logger.Info("gRPC server is running")
		if err := grpcServer.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	if err := apiServer.Run(); err != nil {
		log.Fatal(err)
	}
}
