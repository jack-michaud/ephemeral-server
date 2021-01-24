package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	consul "github.com/hashicorp/consul/api"
	"github.com/jack-michaud/ephemeral-server/bot"
	"github.com/jack-michaud/ephemeral-server/bot/store"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	discordSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	if discordSecret == "" {
		log.Fatalln("DISCORD_CLIENT_SECRET not found in env")
	}
	consulAddress := os.Getenv("CONSUL_ADDRESS")
	if consulAddress == "" {
		log.Fatalln("CONSUL_ADDRESS not found in env")
	}
	if os.Getenv("SECRET_KEY") == "" {
		log.Fatalln("SECRET_KEY not found in env")
	}

	// initialize kv store (consul)
	var kvConn store.IKVStore
	config := consul.DefaultConfig()
	config.Address = consulAddress
	kvConn, err := store.NewKVConsul(config)
	if err != nil {
		log.Fatalln("could not initialize consul:", err)
	}
	err = kvConn.TestLive()
	if err != nil {
		log.Fatalln("could not initialize consul:", err)
	}

	session, err := bot.InitializeBot(ctx, bot.InitializeConfig{
		DiscordSecret: discordSecret,
		KVConn:        kvConn,
	})

	if err != nil {
		log.Fatalln("error initializing bot:", err)
	}
	log.Println("Hello,", session.State.User.Username)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	for {
		select {
		case <-signalChan:
			log.Println("Got close signal")
			cancel()
		case <-ctx.Done():
			log.Println("Closing discord bot session..")
			session.Close()
			log.Println("Closed. Bye!")
			return
		}
	}

}
