package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// Prompts user with list of roles and asks to select one for managing the bot.
func askSetupManagingRoleFunction(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
	Guild, err := s.Guild(m.GuildID)
	possibleRoles := make(discordgo.Roles, 0)
	if err == nil {
		possibleRoles = Guild.Roles
	}
	fields := make([]*discordgo.MessageEmbedField, 0)
	for _, role := range possibleRoles {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  role.Name,
			Value: role.ID,
		})
	}
	s.ChannelMessageSendComplex(
		m.ChannelID,
		&discordgo.MessageSend{
			Content: "You may optionally set up a role that can tell the server to go up or down. By default, anyone can control this.\n`>skip` to skip this step\n`>role <role ID>` (not the name) to set a role",
			Embed: &discordgo.MessageEmbed{
				Fields: fields,
			},
		},
	)
	if len(fields) == 0 {
		s.ChannelMessageSend(
			m.ChannelID,
			"I was going to ask you if you wanted to set up a role, but for some reason I can't get a list of roles. Sorry... Just `>skip`.",
		)
	}
	return nil, nil
}

func handleSetupManagingRoleFunction(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
	roleId := args[1]
	Guild, err := s.Guild(m.GuildID)
	possibleRoles := make(discordgo.Roles, 0)
	if err != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			"Couldn't get roles...Skipping setting the role ID.",
		)
		return nil, nil
	}
	possibleRoles = Guild.Roles
	for _, role := range possibleRoles {
		if roleId == role.ID {
			s.ChannelMessageSend(
				m.ChannelID,
				fmt.Sprintf("Managing Role: %s\nConfirm? `>yes`/`>no`", role.Name),
			)
			config.ManagingRoleId = roleId
			return nil, nil
		}
	}

	msg := "Invalid role, make sure you're using the numeric role ID."
	s.ChannelMessageSend(
		m.ChannelID,
		msg,
	)
	return fmt.Errorf(msg), nil
}
