package main

import (
	"time"

	"github.com/Netflix/go-env"
	"github.com/go-playground/validator"
	"github.com/gorilla/websocket"
	web "github.com/sina-am/social-media/common"
	"github.com/sina-am/social-media/internal/auth/client"
	"go.uber.org/zap"
)

type Settings struct {
	HTTPAddress string `env:"HTTP_ADDRESS,default=localhost:8080"`
	AuthAddress string `env:"AUTH_ADDRESS,default=localhost:5000"`
	MongoURI    string `env:"MONGO_URI,default=mongodb://localhost"`
	MongoDBName string `env:"MONGO_DBNAME,default=chat"`
}

func main() {
	logger, _ := zap.NewProduction()

	var settings Settings
	_, err := env.UnmarshalFromEnviron(&settings)
	if err != nil {
		logger.Fatal(err.Error())
	}

	validate = validator.New()
	storage, err := NewMongoStorage(settings.MongoURI, settings.MongoDBName)
	if err != nil {
		logger.Fatal(err.Error())
	}
	auth := client.NewGRPCClient(settings.AuthAddress)
	service := NewChatService(storage, auth)
	apiServer := APIServer{
		APIServer: web.APIServer{
			Addr:   settings.HTTPAddress,
			Logger: logger,
		},
		Auth:     auth,
		Storage:  storage,
		Service:  service,
		Upgrader: websocket.Upgrader{HandshakeTimeout: 3 * time.Second},
	}

	logger.Fatal(apiServer.Run().Error())
}
