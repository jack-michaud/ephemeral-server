package bot

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jack-michaud/ephemeral-server/bot/serverbridge"
	"github.com/jack-michaud/ephemeral-server/bot/store"
)

var SERVER_TYPES = []string {
  "vanilla-1.16.4",
  "skyfactory-4.2.2",
  "ftb-revelation-3.4.0",
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

func RunEphemeral(ctx context.Context, kvConn store.IKVStore, action EphemeralAction, config *Config, textUpdateChannel chan string) {
  ServerSize := config.Size
  ServerName := fmt.Sprintf("discord-%s", config.ServerId)
  ServerType := config.ServerType
  Region := config.Region
  log.SetPrefix(fmt.Sprintf("[ephemeralctl:%s] ", ServerName))

  if (config.PrivateKey == nil) {
    privateKey, err := GeneratePrivateKey()
    if err != nil {
      textUpdateChannel <- "Unable to generate keys for server. Could not run command :/"
      log.Println("error: unable to generate private key: ", err)
      return
    } else {
      log.Println("Generated new private keys")
    }

    config.PrivateKey = privateKey
    config.SaveConfig(kvConn)
  }

  // Write public and private key to local cache for ansible
  keysDir := fmt.Sprintf(".cache/config-%s/keys", ServerName)
  err := os.MkdirAll(keysDir, 0700)
  if err != nil {
    log.Println("Could not make keys directory:", err)
    textUpdateChannel <- "Unable to write key locally. Could not run command :/"
    return
  }
  err = ioutil.WriteFile(
    fmt.Sprintf("%s/minecraft-%s", keysDir, ServerName),
    GetPrivateKeyString(config.PrivateKey),
    0700,
  )
  if err != nil {
    log.Println("Could not write public key file:", err)
    textUpdateChannel <- "Unable to save public key. Could not run command :/"
    return
  }
  publicKey, err := GetAuthorizedFilePublicKeyString(config.PrivateKey)
  if err != nil {
    log.Println("Could not write public key to string:", err)
    textUpdateChannel <- "Unable to write key to server. Could not run command :/"
    return
  }
  err = ioutil.WriteFile(
    fmt.Sprintf("%s/minecraft-%s.pub", keysDir, ServerName),
    publicKey,
    0700,
  )
  if err != nil {
    log.Println("Could not write public key file:", err)
    textUpdateChannel <- "Unable to save public key. Could not run command :/"
    return
  }

  shutdownMinecraftServer := func() {
    sshClient, err := serverbridge.ConnectToServer(ctx, &serverbridge.ConnectOptions{
      ServerIpAddress: config.ServerIpAddress,
      PrivateKey: config.PrivateKey,
    }, kvConn)
    if err == nil {
      serverbridge.ShutdownMinecraftServer(ctx, sshClient)
    } else {
      log.Println(err)
    }
  }


  var Env []string = os.Environ()
  var Args []string = make([]string, 0)
  if (config.CloudProvider == "aws") {
    Creds := config.AwsCreds
    Env = append(
      Env,
      "CLOUD_PROVIDER=aws",
      fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", Creds.SecretAccessKey),
      fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", Creds.AccessKeyId),
      fmt.Sprintf("AWS_DEFAULT_REGION=%s", config.Region),
    )
  } else if (config.CloudProvider == "digitalocean") {
    Creds := config.DigitalOceanCreds
    Env = append(
      Env,
      "CLOUD_PROVIDER=digitalocean",
      fmt.Sprintf("DIGITALOCEAN_TOKEN=%s", Creds.AccessKey),
    )
  }

  Args = append(Args,
    "-t", ServerType,
    "-n", ServerName,
    "-s", ServerSize,
    "-r", Region,
  )

  var actionFlag string
  var actionString string
  switch action {
  case NO_OP:
  case DESTROY_VPC:
    actionString = "Destroying server"
    actionFlag = "-d"
    shutdownMinecraftServer()
  case DESTROY_ALL:
    actionString = "Destroying server and persistent device"
    actionFlag = "-D"
    shutdownMinecraftServer()
  case CREATE:
    actionString = "Starting to create server"
    actionFlag = "-c"
  case GET_IP:
    actionString = "Getting IP"
    actionFlag = "-i"
    textUpdateChannel <- *config.ServerIpAddress
    return
  case ANSIBLE_PROVISION:
    actionString = "Reinstalling software..."
    actionFlag = "-I"
  }

  EPHEMERAL_BIN := "./ephemeralctl.sh"
  cmd := exec.CommandContext(ctx, EPHEMERAL_BIN, append(Args, actionFlag)...)
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
  // This goroutine will:
  // - Read output
  // - Update discord channel
  // - Set IP address
  go func() {
    scanner := bufio.NewScanner(stdoutPipe)
    for scanner.Scan() {
      line := scanner.Text()
      log.Println(line)
      if action == CREATE {
        if strings.Contains(line, "Successfully applied terraform") {
          textUpdateChannel <- fmt.Sprintf("Successfully created VPS! Starting %s server (May take a couple minutes.)", ServerType)
        }
        if strings.Contains(line, "Failed to apply terraform") {
          log.Println("ERROR FAILED TO CREATE VPS")
          textUpdateChannel <- "Failed to create VPS"
        }
        if strings.Contains(line, "Successfully applied ansible") {
          ip, err := exec.Command(EPHEMERAL_BIN, append(Args, "-i")...).Output()
          ipString := strings.TrimSpace(string(ip))
          if err != nil {
            log.Println("Tried to get IP, but failed:", err)
            textUpdateChannel <- fmt.Sprintf("Successfully created %s server!", ServerType)
          } else {
            textUpdateChannel <- fmt.Sprintf("Successfully created %s server! IP: %s:25565", ServerType, ipString)
            config.ServerIpAddress = &ipString
            config.SaveConfig(kvConn)
          }
        }
        if strings.Contains(line, "Failed to apply ansible") {
          log.Println("ERROR FAILED TO CREATE VPS")
          textUpdateChannel <- "Failed to create VPS"
        }
      }
      if action == DESTROY_VPC {
        if strings.Contains(line, "Destroy complete!") {
          textUpdateChannel <- fmt.Sprintf("Shut down server.")
          config.ServerIpAddress = nil
          config.SaveConfig(kvConn)
        }
      }
      if action == DESTROY_ALL {
        if strings.Contains(line, "Destroy complete!") {
          textUpdateChannel <- fmt.Sprintf(
            "Shut down server and deleted persistent data volume.",
          )
          config.ServerIpAddress = nil
          config.SaveConfig(kvConn)
        }
      }
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
  textUpdateChannel <- fmt.Sprintf(
    "%s...(Please be patient, it may take a couple minutes!)",
    actionString,
  )
  err = cmd.Wait()
  if err != nil {
    textUpdateChannel <- "Failed :/"
    log.Println("failed:", err)
  }
}
