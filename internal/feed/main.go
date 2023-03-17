package main

import (
	"log"

	env "github.com/Netflix/go-env"
	"github.com/go-playground/validator"
	web "github.com/sina-am/social-media/common"
	"github.com/sina-am/social-media/internal/auth/client"
)

type Settings struct {
	HTTPAddress string `env:"HTTP_ADDRESS,default=:8080"`
	AuthAddress string `env:"AUTH_ADDRESS,default=localhost:5000"`
	MongoURI    string `env:"MONGO_URI,default=mongodb://localhost"`
	MongoDBName string `env:"MONGO_DBNAME,default=feeds"`
}

func main() {
	var settings Settings
	_, err := env.UnmarshalFromEnviron(&settings)
	if err != nil {
		log.Fatal(err)
	}

	validate = validator.New()
	storage, err := NewMongoStorage(settings.MongoURI, settings.MongoDBName)
	if err != nil {
		log.Fatal(err)
	}

	apiServer := APIServer{
		APIServer: web.APIServer{
			Addr: settings.HTTPAddress,
		},
		Auth:    client.NewGRPCClient(settings.AuthAddress),
		Storage: storage,
	}
	log.Fatal(apiServer.Run())
}
