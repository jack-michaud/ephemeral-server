package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
)

func SendError(s *discordgo.Session, channelId string, err string) {
  s.ChannelMessageSend(channelId, fmt.Sprintf("error: %s", err))
}

type InitializeConfig struct {
  DiscordSecret string
  RedisConn *redis.Client
}

func InitializeBot(ctx context.Context, initializeConfig InitializeConfig) (*discordgo.Session, error) {
  discordSecret := initializeConfig.DiscordSecret
  redisConn := initializeConfig.RedisConn
  if discordSecret == "" {
    return nil, fmt.Errorf("Must provide discordSecret")
  }
  if redisConn == nil {
    return nil, fmt.Errorf("Must provide redisConn")
  }

  session, err := discordgo.New("Bot " + discordSecret)
  if err != nil {
    return nil, fmt.Errorf("error initializing bot: %s", err)
  }

  session.UserAgent = "EphemeralServer (https://github.com/jack-michaud/ephemeral-server)"

  configMap := NewConfigMap()
  configStateMachine := NewConfigStateMachine()

  session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
    ID := g.ID
    log.Println("Got guild create event for guild ID:", ID, ". (", g.Guild.Name, ")")
    config, err := GetConfigForServerId(ID, redisConn)
    if err != nil {
      log.Println("Could not fetch config from store")
    }
    configMap.Set(ID, *config)
  })

  session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
    GuildID := m.GuildID
    _, err := s.Guild(m.GuildID)
    if err != nil {
      log.Println("Could not get guild", GuildID, "from guildmap")
      return
    }
    config, exists := configMap.Get(GuildID)
    if !exists {
      SendError(
        s,
        m.ChannelID,
        "Could not find config, initialized or not. This could be an issue with the bot.",
      )
    }

    // Only execute state machine if message comes from a authorized user
    // If managing Role Id is empty, anyone can execute the state machine. 
    if (
      m.Author.ID != s.State.User.ID &&
      (config.ManagingRoleId == "" ||
      contains(m.Member.Roles, config.ManagingRoleId))) {

      state, found := configStateMachine.Get(GuildID)
      if !found {
        state = InitializeConfigStateMachine(redisConn)
      }
      configStateMachine.Set(GuildID, state.GetNextStateFromMessage(s, m, &config))
      configMap.Set(GuildID, config)
    }
  })

  session.Identify.Intents = discordgo.MakeIntent(
    discordgo.IntentsAllWithoutPrivileged,
  )

  err = session.Open()
  if err != nil {
    return nil, fmt.Errorf("error initializing bot: %s", err)
  }

  return session, nil
}
