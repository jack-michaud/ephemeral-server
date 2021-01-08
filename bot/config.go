package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type AwsCreds struct {
  AccessKeyId string
  SecretAccessKey string
}
type DigitalOceanCreds struct {
  AccessKey string
}

type PrivateKey = []byte

type Config struct {
  CloudProvider string
  DigitalOceanCreds *DigitalOceanCreds
  Region string
  Size string
  // Aws.
  AwsCreds *AwsCreds
  // Private key to access the VPS.
  PrivateKey *PrivateKey
  ManagingRoleId string
}

func GetConfigForServerId(Id string) (Config, error) {
  return Config{
    CloudProvider: "",
    DigitalOceanCreds: &DigitalOceanCreds{
      AccessKey: "123",
    },
    AwsCreds: nil,
    PrivateKey: nil,
  }, nil
}

func (c *Config) SaveConfig() error {
  return nil
}

type ConfigMap struct {
  rwmap sync.Map
}

func NewConfigMap() *ConfigMap {
  return &ConfigMap{
    rwmap: sync.Map{},
  }
}

func (cm *ConfigMap) Get(key string) (Config, bool) {
  data, found := cm.rwmap.Load(key)
  if !found {
    return Config{}, found
  }
  return data.(Config), found
}

func (cm *ConfigMap) Set(key string, config Config) {
  cm.rwmap.Store(key, config)
}

type Action = func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error
var NilAction = func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
  return nil
}

type State struct {
  id string
  Action Action
  // Message regexps that trigger the next state
  nextStates map[string] State
}

func (s *State) GetNextStateFromMessage(session *discordgo.Session, message *discordgo.MessageCreate, config *Config) State {
  messageContent := message.Content
  for pattern, nextState := range s.nextStates {
    rg, err := regexp.Compile(pattern)
    if err != nil {
      log.Println("could not compile state regex", s.id, ":", err)
    }

    strings := rg.FindAllStringSubmatch(messageContent, -1)
    if strings != nil {
      log.Println("Found matching state:", nextState.id)
      err := nextState.Action(session, message, config, strings[0])
      if err != nil {
        log.Println("error running state action:", err)
        return *s
      }
      return nextState
    }
  }
  return *s
}

func (s *State) AddState(pattern string, newState State) {
  s.nextStates[pattern] = newState
}

type ConfigStateMachine struct {
  rwmap sync.Map
}

func NewConfigStateMachine() *ConfigStateMachine {
  return &ConfigStateMachine{
    rwmap: sync.Map{},
  }
}

func (cm *ConfigStateMachine) Get(key string) (State, bool) {
  data, found := cm.rwmap.Load(key)
  if !found {
    return State{}, found
  }
  return data.(State), found
}

func (cm *ConfigStateMachine) Set(key string, state State) {
  cm.rwmap.Store(key, state)
}


func NewState(id string, action Action) State {
  nextStates := make(map[string] State)
  return State{
    id: id,
    Action: action,
    nextStates: nextStates,
  }
}

const DATA_SECURITY_NOTE = "(Note: All credentials are encrypted in transit and at rest)"
func InitializeConfigStateMachine() State {
  configRoot := NewState(
    "root-config-flow",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      s.ChannelMessageSend(
        m.ChannelID,
        "Ready to configure EphemeralServer? (Note: requires sensitive information, you should do this in a private channel) (`>cancel`/`>continue`)",
      )
      return nil
    },
  )
  cancelStep := NewState(
    "cancel",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      s.ChannelMessageSend(
        m.ChannelID,
        "Canceling configuration.",
      )
      return nil
    },
  )

  // Ask and handle cloud provider
  askCloudProviderStep := NewState(
    "ask-cloud-provider",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf(
          "Great, thanks %s. You can cancel the setup at any time with `>cancel`. Firstly, what cloud provider are you using? (`>provider digitalocean` (recommended, it's cheaper) or `>provider aws`)",
          m.Author.Username,
        ),
      )
      return nil
    },
  )
  cloudProviderStep := NewState(
    "handle-cloud-provider",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      cloudProvider := args[1]
      config.CloudProvider = cloudProvider
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf(
          "Set provider: `%s`. Is that right? (`>yes`/`>no`)",
          cloudProvider,
        ),
      )
      return nil
    },
  )

  askCloudCredentials := NewState(
    "ask-cloud-credentials",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      msg := ""
      if (config.CloudProvider == "digitalocean") {
        msg = "For digitalocean, all you need is a Personal Access Token. For more info, go here: https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/. Input your personal access token with `>keys <token>`. For example, `>keys 123123214135234`"
      } else if (config.CloudProvider == "aws") {
        msg = fmt.Sprintf("For AWS, you need to create an access key with for an IAM user *in the %s region* with at least `ec2:*` permissions. For more info on creating access keys and IAM users, go here: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html. Input your access key ID and secret access key with `>keys <access_key_id> <secret_access_key>`. For example, `>keys AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`", config.Region)
      } else {
        return fmt.Errorf("Invalid cloud provider: %s", config.CloudProvider)
      }
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf("%s %s", msg, DATA_SECURITY_NOTE),
      )
      return nil
    },
  )
  handleCloudCredentials := NewState(
    "handle-cloud-credentials",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      cloudProvider := config.CloudProvider
      if cloudProvider == "digitalocean" {
        AccessKey := args[1]
        config.DigitalOceanCreds = &DigitalOceanCreds{
          AccessKey,
        }
        s.ChannelMessageSend(
          m.ChannelID,
          fmt.Sprintf(
            "Private Access Key: %s\nLook right? (`>yes`,`>no`)",
            AccessKey,
          ),
        )

      }
      if cloudProvider == "aws" {
        AccessKeyId := args[1]
        SecretAccessKey := args[2]
        config.AwsCreds = &AwsCreds{
          AccessKeyId,
          SecretAccessKey,
        }
        s.ChannelMessageSend(
          m.ChannelID,
          fmt.Sprintf(
            "Access Key ID: %s\nSecret Access Key:%s\nLook right? (`>yes`,`>no`)",
            AccessKeyId,
            SecretAccessKey,
          ),
        )
      }
      return nil
    },
  )

  askRegion := NewState(
    "ask-region",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      msg := "What region would you like to run the server in?"
      if config.CloudProvider == "digitalocean" {
        msg = fmt.Sprintf(
          "%s (%s)\nAvailable regions: %s",
          msg,
          "To set region, use `>region <region>`, e.g. `>region nyc1`.",
          strings.Join(DIGITALOCEAN_VALID_REGIONS, ","),
        )
      } else if config.CloudProvider == "aws" {
        msg = fmt.Sprintf(
          "%s (%s)\nAvailable regions: %s",
          msg,
          "To find a list of all AWS regions, go here: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html. To set region, use `>region <region>`, e.g. `>region us-east-1`. Code will assume availability zone `<region>a`.",
          strings.Join(AWS_VALID_REGIONS, ","),
        )
      } else {
        msg = "Cloud provider not set or invalid"
        s.ChannelMessageSend(
          m.ChannelID,
          msg,
        )
        return fmt.Errorf(msg)
      }

      s.ChannelMessageSend(
        m.ChannelID,
        msg,
      )

      return nil
    },
  )
  handleRegion := NewState(
    "handle-region",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      region := args[1]
      if config.CloudProvider == "digitalocean" {
        if !contains(DIGITALOCEAN_VALID_REGIONS, region) {
          msg := fmt.Sprintf("Error: region %s is not a valid Digital Ocean region", region)
          s.ChannelMessageSend(
            m.ChannelID,
            msg,
          )
          return fmt.Errorf(msg)
        }
      }
      if config.CloudProvider == "aws" {
        if !contains(AWS_VALID_REGIONS, region) {
          msg := fmt.Sprintf("Error: region %s is not a valid AWS region", region)
          s.ChannelMessageSend(
            m.ChannelID,
            msg,
          )
          return fmt.Errorf(msg)
        }
      }

      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf("Region: %s. Look good? (`>yes`/`>no`)", region),
      )
      config.Region = region
      return nil
    },
  )

  askSize := NewState(
    "ask-size",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      return nil
    },
  )
  handleSize := NewState(
    "handle-size",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      return nil
    },
  )

  askSetupManagingRole := NewState(
    "ask-setup-managing-role",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      return nil
    },
  )
  handleSetupManagingRole := NewState(
    "handle-setup-managing-role",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      return nil
    },
  )

  saveConfig := NewState(
    "save-config",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      err := config.SaveConfig()
      if err != nil {
        s.ChannelMessageSend(
          m.ChannelID,
          fmt.Sprintf("Unable to save config: %s", err),
        )
      } else {
        s.ChannelMessageSend(
          m.ChannelID,
          fmt.Sprintf("Saved config. %s", DATA_SECURITY_NOTE),
        )
      }
      serialize, _ := json.Marshal(*config)
      fmt.Println(string(serialize))
      return nil
    },
  )

  rootState := NewState("root", NilAction)
  rootState.AddState(`^>eph(emeral)* config$`, configRoot)

  configRoot.AddState(`>cancel`, cancelStep)
  configRoot.AddState(`>continue`, askCloudProviderStep)

  askCloudProviderStep.AddState(`>cancel`, cancelStep)
  askCloudProviderStep.AddState(`>provider (digitalocean|aws)`, cloudProviderStep)

  cloudProviderStep.AddState(`>no`, askCloudProviderStep)
  cloudProviderStep.AddState(`>cancel`, cancelStep)
  cloudProviderStep.AddState(`>yes`, askRegion)

  askRegion.AddState(`>cancel`, cancelStep)
  askRegion.AddState(`>region (.*)`, handleRegion)

  handleRegion.AddState(`>no`, askRegion)
  handleRegion.AddState(`>cancel`, cancelStep)
  handleRegion.AddState(`>yes`, askCloudCredentials)

  askCloudCredentials.AddState(`>cancel`, cancelStep)
  askCloudCredentials.AddState(`>keys (.*) (.*)`, handleCloudCredentials)
  askCloudCredentials.AddState(`>keys (.*)`, handleCloudCredentials)

  handleCloudCredentials.AddState(`>no`, askCloudCredentials)
  handleCloudCredentials.AddState(`>cancel`, cancelStep)
  handleCloudCredentials.AddState(`>yes`, askSize)

  askSize.AddState(`>size (.*)`, handleSize)
  askSize.AddState(`>cancel`, cancelStep)

  handleSize.AddState(`>no`, askSize)
  handleSize.AddState(`>cancel`, cancelStep)
  handleSize.AddState(`>yes`, askSetupManagingRole)

  askSetupManagingRole.AddState(`>skip`, saveConfig)
  askSetupManagingRole.AddState(`>cancel`, cancelStep)
  askSetupManagingRole.AddState(`>role (.*)`, handleSetupManagingRole)


  cancelStep.AddState(`^>ephemeral config.*`, configRoot)

  return rootState
}

