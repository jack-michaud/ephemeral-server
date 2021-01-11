package bot

import (
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

func RunEphemeral(ctx context.Context, action EphemeralAction, config *Config) {
  //log.SetPrefix("[ephemeralctl] ")
  ServerSize := config.Size
  ServerName := fmt.Sprintf("discord-%s", config.ServerId)
  ServerType := config.ServerType
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
    fmt.Sprintf("-t %s", ServerType),
    fmt.Sprintf("-n %s", ServerName),
    fmt.Sprintf("-s %s", ServerSize),
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

  cmd := exec.CommandContext(ctx, "/home/jack/Code/minecraft-server-bot/ephemeralctl.sh", Args...)
  cmd.Env = Env

  log.Println("launch:", Args)
  log.Println("launch:", cmd.String())

  log.Println("Starting eph command")
  err := cmd.Start()
  if err != nil {
    log.Println("failed:", err)
  }
  log.Println("waiting...")
  err = cmd.Wait()
  log.Println("done waiting")
  output, err := cmd.CombinedOutput()
  if err != nil {
    log.Println("failed:", err)
  } else {
    log.Println(string(output))
  }
}
