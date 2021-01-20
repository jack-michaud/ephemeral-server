package serverbridge

import (
	"bufio"
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"
)

func RunCommand(ctx context.Context, client *ssh.Client, command string) (stdOutLines chan string, stdErrLines chan string, doneChan chan bool, err error) {
	stdOutLines = make(chan string, 1)
	stdErrLines = make(chan string, 1)
	doneChan = make(chan bool, 1)
	sess, err := client.NewSession()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create new session: %s", err)
	}

	// Read stdout
	go func() {
		stdout, _ := sess.StdoutPipe()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			stdOutLines <- line
		}
	}()

	// Read stderr
	go func() {
		stderr, _ := sess.StderrPipe()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stdErrLines <- line
		}
	}()

	err = sess.Start(command)
	log.Println("Running command:", command)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start command: %s", err)
	}

	// stop running if context ends
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Closing command...")
				return
			}
		}
	}()

	go func() {
		sess.Wait()
		defer sess.Close()
		doneChan <- true
	}()

	return stdOutLines, stdErrLines, doneChan, nil
}
