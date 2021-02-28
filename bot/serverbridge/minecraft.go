package serverbridge

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func ShutdownMinecraftServer(ctx context.Context, client *ssh.Client) error {
	cmd := "sudo systemd stop mc-server.service"
	stdOutLines, _, doneChan, err := RunCommand(ctx, client, cmd)
	if err == nil {
		return fmt.Errorf("could not stop server command: %s", err)
	}
	go func() {
		for line := range stdOutLines {
			log.Println(line)
		}
	}()

	timer := time.NewTimer(time.Second * 10)
	select {
	case <-doneChan:
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}
}

func RestartMinecraftServer(ctx context.Context, client *ssh.Client) error {
	cmd := "sudo systemd restart mc-server.service"
	stdOutLines, _, doneChan, err := RunCommand(ctx, client, cmd)
	if err != nil {
		return fmt.Errorf("could not run restart server command: %s", err)
	}
	go func() {
		for line := range stdOutLines {
			log.Println(line)
		}
	}()

	timer := time.NewTimer(time.Second * 30)
	select {
	case <-doneChan:
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}
}
