package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	consul "github.com/hashicorp/consul/api"
	"github.com/jack-michaud/ephemeral-server/bot"
	bridge "github.com/jack-michaud/ephemeral-server/bot/serverbridge"
	"github.com/jack-michaud/ephemeral-server/bot/store"
)

func main() {

	ServerId := flag.String("serverId", "", "Specifies the server ID of the VPS to connect to")
	cmd := flag.String("cmd", "", "command to run on VPS")
	flag.Parse()

	if len(*ServerId) == 0 {
		os.Exit(1)
	}

	consulAddress := os.Getenv("CONSUL_ADDRESS")
	if consulAddress == "" {
		log.Fatalln("CONSUL_ADDRESS not found in env")
	}
	if os.Getenv("SECRET_KEY") == "" {
		log.Fatalln("SECRET_KEY not found in env")
	}

	// initialize kv store (consul)
	consulConfig := consul.DefaultConfig()
	consulConfig.Address = consulAddress
	var kvConn store.IKVStore
	kvConn, err := store.NewKVConsul(consulConfig)

	defer kvConn.Cleanup()
	if err != nil {
		log.Fatalln("could not initialize consul:", err)
	}
	err = kvConn.TestLive()
	if err != nil {
		log.Fatalln("could not initialize consul:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, err := bot.GetConfigForServerId(*ServerId, kvConn)
	if err != nil {
		log.Fatalln("could not fetch config:", err)
	}
	b, _ := json.Marshal(config)
	fmt.Println(string(b))

	client, err := bridge.ConnectToServer(ctx, &bridge.ConnectOptions{
		PrivateKey:      config.PrivateKey,
		ServerIpAddress: config.ServerIpAddress,
	}, kvConn)
	if err != nil {
		log.Fatalln("unable to make connection:", err)
	} else {
		log.Println("connected! :)")
	}
	defer func() {
		err = client.Close()
		if err != nil {
			log.Fatalln("unable to close connection:", err)
		}
	}()

	var stdOutLines chan string
	if *cmd == "" {
		stdOutLines, _, _, err = bridge.RunCommand(ctx, client, "sudo journalctl -f -u mc-server.service")
	} else {
		stdOutLines, _, _, err = bridge.RunCommand(ctx, client, *cmd)
	}
	if err != nil {
		log.Fatalln("could not run command:", err)
	} else {
		log.Println("running", cmd)
	}

	go func() {
		for line := range stdOutLines {
			log.Println(line)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	for {
		select {
		case <-signalChan:
			log.Println("Got close signal")
			if err != nil {
				log.Println(err)
			}
			return
		case <-ctx.Done():
			return
		}
	}
}
