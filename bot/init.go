package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func SendError(s *discordgo.Session, channelId string, err string) {
  s.ChannelMessageSend(channelId, fmt.Sprintf("error: %s", err))
}

func InitializeBot(ctx context.Context, discordSecret string) (*discordgo.Session, error) {
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
    config, err := GetConfigForServerId(ID)
    if err != nil {
      log.Println("Could not fetch config from store")
    }
    configMap.Set(ID, config)
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
        state = InitializeConfigStateMachine()
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
