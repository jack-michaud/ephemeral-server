package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jack-michaud/ephemeral-server/bot/store"
)

func SendError(s *discordgo.Session, channelId string, err string) {
	s.ChannelMessageSend(channelId, fmt.Sprintf("error: %s", err))
}

type InitializeConfig struct {
	DiscordSecret string
	KVConn        store.IKVStore
}

func InitializeBot(ctx context.Context, initializeConfig InitializeConfig) (*discordgo.Session, error) {
	discordSecret := initializeConfig.DiscordSecret
	kvConn := initializeConfig.KVConn
	if discordSecret == "" {
		return nil, fmt.Errorf("Must provide discordSecret")
	}
	if kvConn == nil {
		return nil, fmt.Errorf("Must provide kvConn")
	}

	session, err := discordgo.New("Bot " + discordSecret)
	if err != nil {
		return nil, fmt.Errorf("error initializing bot: %s", err)
	}

	session.UserAgent = "EphemeralServer (https://github.com/jack-michaud/ephemeral-server)"

	configStateMachine := NewConfigStateMachine()

	session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		ID := g.ID
		log.Println("Got guild create event for guild ID:", ID, ". (", g.Guild.Name, ")")
	})

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		GuildID := m.GuildID
		_, err := s.Guild(m.GuildID)
		if err != nil {
			log.Println("Could not get guild", GuildID, "from guildmap")
			return
		}
		config, err := GetConfigForServerId(GuildID, kvConn)
		if err != nil {
			SendError(
				s,
				m.ChannelID,
				fmt.Sprintf("Could not get config: %s", err),
			)
		}

		// Only execute state machine if message comes from a authorized user
		// If managing Role Id is empty, anyone can execute the state machine.
		if m.Author.ID != s.State.User.ID &&
			(config.ManagingRoleId == "" ||
				contains(m.Member.Roles, config.ManagingRoleId)) {

			state, found := configStateMachine.Get(GuildID)
			if !found {
				state = InitializeConfigStateMachine(ctx, kvConn)
			}
			configStateMachine.Set(GuildID, state.GetNextStateFromMessage(s, m, config))
			config.SaveConfig(kvConn)
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
