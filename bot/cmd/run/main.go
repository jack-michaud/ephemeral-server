package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-redis/redis"
	"github.com/jack-michaud/ephemeral-server/bot"
)

func main()  {
  ctx, cancel := context.WithCancel(context.Background())
  discordSecret := os.Getenv("DISCORD_CLIENT_SECRET")
  if (discordSecret == "") {
    log.Fatalln("DISCORD_CLIENT_SECRET not found in env")
  }
  redisUrl := os.Getenv("REDIS_URL")
  if (redisUrl == "") {
    log.Fatalln("REDIS_URL not found in env")
  }

  // initialize redis
  options, err := redis.ParseURL(redisUrl)
  if err != nil {
    log.Fatalln("could not initialize redis options:", err)
  }

  conn := redis.NewClient(options)
  ret := conn.Ping()
  retVal := ret.Val()
  if retVal != "PONG" {
    log.Fatalln("could not ping redis: got ping response:", retVal)
  }
  if err != nil {
    log.Fatalln("could not initialize redis:", err)
  }
  session, err := bot.InitializeBot(ctx, bot.InitializeConfig{
    DiscordSecret: discordSecret,
    RedisConn: conn,
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
