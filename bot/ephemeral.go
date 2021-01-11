package bot

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
)

var SERVER_TYPES = []string {
  "vanilla-1.16.4",
}

// Ephemeralctl actions
type EphemeralAction int
const (
  NO_OP EphemeralAction = iota
  DESTROY_ALL
  DESTROY_VPC
  // Idempotent
  CREATE
  GET_IP
  ANSIBLE_PROVISION
)

func RunEphemeral(ctx context.Context, action EphemeralAction, config *Config, textUpdateChannel chan string) {
  ServerSize := config.Size
  ServerName := fmt.Sprintf("discord-%s", config.ServerId)
  ServerType := config.ServerType
  log.SetPrefix(fmt.Sprintf("[ephemeralctl:%s] ", ServerName))

  var Env []string = os.Environ()
  var Args []string = make([]string, 0)
  if (config.CloudProvider == "aws") {
    Creds := config.AwsCreds
    Env = append(
      Env,
      "CLOUD_PROVDER=aws",
      fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", Creds.SecretAccessKey),
      fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", Creds.AccessKeyId),
      fmt.Sprintf("AWS_DEFAULT_REGION=%s", config.Region),
    )
  } else if (config.CloudProvider == "digitalocean") {
    Creds := config.DigitalOceanCreds
    Env = append(
      Env,
      "CLOUD_PROVDER=digitalocean",
      fmt.Sprintf("DIGITAL_OCEAN_TOKEN=%s", Creds.AccessKey),
    )
  }

  Args = append(Args,
    "-t", ServerType,
    "-n", ServerName,
    "-s", ServerSize,
  )

  var actionFlag string
  switch action {
  case NO_OP:
  case DESTROY_VPC:
    actionFlag = "-d"
  case DESTROY_ALL:
    actionFlag = "-D"
  case CREATE:
    actionFlag = "-c"
  case GET_IP:
    actionFlag = "-i"
  case ANSIBLE_PROVISION:
    actionFlag = "-I"
  }

  Args = append(Args, actionFlag)

  cmd := exec.CommandContext(ctx, "./ephemeralctl.sh", Args...)
  cmd.Env = Env

  log.Println("launch:", Args)
  log.Println("launch:", cmd.String())
  stdoutPipe, err := cmd.StdoutPipe()
  if err != nil {
    log.Println("error: could not get stdoutpipe:", err)
  }
  stderrPipe, err := cmd.StderrPipe()
  if err != nil {
    log.Println("error: could not get stderrpipe:", err)
  }

  cmd.Start()
  // Read output
  go func() {
    scanner := bufio.NewScanner(stdoutPipe)
    for scanner.Scan() {
      line := scanner.Text()
      log.Println(line)
    }
  }()

  // Read output
  go func() {
    scanner := bufio.NewScanner(stderrPipe)
    for scanner.Scan() {
      line := scanner.Text()
      log.Println(line)
    }
  }()

  // Wait for command to finish
  err = cmd.Wait()
  if err != nil {
    log.Println("failed:", err)
  }
}
