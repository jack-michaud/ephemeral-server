package serverbridge

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/crypto/ssh"
)

// java[15054]: [09:33:29] [RCON Client /127.0.0.1 #8/INFO]: Thread RCON Client /127.0.0.1 shutting down
// java[15054]: [09:33:36] [RCON Listener #1/INFO]: Thread RCON Client /127.0.0.1 started
// java[15054]: [09:33:36] [Server thread/INFO]: [Rcon] <Lomz> Hiiii
// java[15054]: [09:33:36] [RCON Client /127.0.0.1 #9/INFO]: Thread RCON Client /127.0.0.1 shutting down
// java[15054]: [09:34:47] [RCON Listener #1/INFO]: Thread RCON Client /127.0.0.1 started
// java[15054]: [09:34:47] [Server thread/INFO]: [Rcon] <Lomz> Hiiii
// java[15054]: [09:34:47] [RCON Client /127.0.0.1 #10/INFO]: Thread RCON Client /127.0.0.1 shutting down
// java[15054]: [09:34:56] [RCON Listener #1/INFO]: Thread RCON Client /127.0.0.1 started
// java[15054]: [09:34:56] [Server thread/INFO]: [Rcon] <Lomz> Hiiii
// java[15054]: [09:34:56] [RCON Client /127.0.0.1 #11/INFO]: Thread RCON Client /127.0.0.1 shutting down
// java[15054]: [09:35:19] [Server thread/INFO]: <_lomz_> test

type Bridge struct {
  sshClient *ssh.Client
  discordSession *discordgo.Session
}

func (b *Bridge) SendToMinecraft(message string) error {
  return nil
}

func (b *Bridge) SendToDiscord(message string) error {
  return nil
}
