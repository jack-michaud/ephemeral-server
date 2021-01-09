package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/jack-michaud/ephemeral-server/bot"
)

func main()  {
  ctx, cancel := context.WithCancel(context.Background())
  discordSecret := os.Getenv("DISCORD_CLIENT_SECRET")
  if (discordSecret == "") {
    log.Fatalln("DISCORD_CLIENT_SECRET not found in env")
  }

  session, err := bot.InitializeBot(ctx, discordSecret)
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
