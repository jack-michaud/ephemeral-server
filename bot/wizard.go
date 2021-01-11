package bot

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jack-michaud/ephemeral-server/bot/store"
)

const ISSUE_LINK = "https://github.com/jack-michaud/ephemeral-server/issues/new"

// Returns an error and optionally a state that should be next.
type Action = func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State)
var NilAction = func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
  return nil, nil
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
      log.Println(message.GuildID, "state:", nextState.id)
      err, forcedNextState := nextState.Action(session, message, config, strings[0])
      if err != nil {
        log.Println(message.GuildID, "error running state action:", err)
        if forcedNextState != nil {
          return *forcedNextState
        }
        return *s
      }
      if forcedNextState != nil {
        return *forcedNextState
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
func InitializeConfigStateMachine(ctx context.Context, kvConn store.IKVStore) State {
  rootState := NewState("root", NilAction)

  configRoot := NewState(
    "root-config-flow",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      s.ChannelMessageSend(
        m.ChannelID,
        "Ready to configure EphemeralServer? (Note: requires sensitive information, you should do this in a private channel) (`>cancel`/`>continue`)",
      )
      return nil, nil
    },
  )
  cancelStep := NewState(
    "cancel",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      s.ChannelMessageSend(
        m.ChannelID,
        "Canceling configuration.",
      )
      return nil, &rootState
    },
  )

  // Ask and handle cloud provider
  askCloudProviderStep := NewState(
    "ask-cloud-provider",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf(
          "Great, thanks %s. You can cancel the setup at any time with `>cancel`. Firstly, what cloud provider are you using? (`>provider digitalocean` (recommended, it's cheaper) or `>provider aws`)",
          m.Author.Username,
        ),
      )
      return nil, nil
    },
  )
  cloudProviderStep := NewState(
    "handle-cloud-provider",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      cloudProvider := args[1]
      config.CloudProvider = cloudProvider
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf(
          "Set provider: `%s`. Is that right? (`>yes`/`>no`)",
          cloudProvider,
        ),
      )
      return nil, nil
    },
  )

  askCloudCredentials := NewState(
    "ask-cloud-credentials",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      msg := ""
      if (config.CloudProvider == "digitalocean") {
        msg = "For digitalocean, all you need is a Personal Access Token. For more info, go here: https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/. Input your personal access token with `>keys <token>`. For example, `>keys 123123214135234`"
      } else if (config.CloudProvider == "aws") {
        msg = fmt.Sprintf("For AWS, you need to create an access key with for an IAM user *in the %s region* with at least `ec2:*` permissions. For more info on creating access keys and IAM users, go here: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html. Input your access key ID and secret access key with `>keys <access_key_id> <secret_access_key>`. For example, `>keys AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`", config.Region)
      } else {
        return fmt.Errorf("Invalid cloud provider: %s", config.CloudProvider), nil
      }
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf("%s %s", msg, DATA_SECURITY_NOTE),
      )
      return nil, nil
    },
  )
  handleCloudCredentials := NewState(
    "handle-cloud-credentials",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
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
            "Access Key ID: %s\nSecret Access Key: %s\nLook right? (`>yes`,`>no`)",
            AccessKeyId,
            SecretAccessKey,
          ),
        )
      }
      return nil, nil
    },
  )

  askRegion := NewState(
    "ask-region",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      msg := "What region would you like to run the server in?"
      if config.CloudProvider == "digitalocean" {
        msg = fmt.Sprintf(
          "%s (%s)\nAvailable regions: %s",
          msg,
          "To set region, use `>region <region>`, e.g. `>region nyc1`.",
          strings.Join(DIGITALOCEAN_VALID_REGIONS, ", "),
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
        return fmt.Errorf(msg), nil
      }

      s.ChannelMessageSend(
        m.ChannelID,
        msg,
      )

      return nil, nil
    },
  )
  handleRegion := NewState(
    "handle-region",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      region := args[1]
      if config.CloudProvider == "digitalocean" {
        if !contains(DIGITALOCEAN_VALID_REGIONS, region) {
          msg := fmt.Sprintf("Error: region %s is not a valid Digital Ocean region", region)
          s.ChannelMessageSend(
            m.ChannelID,
            msg,
          )
          return fmt.Errorf(msg), nil
        }
      }
      if config.CloudProvider == "aws" {
        if !contains(AWS_VALID_REGIONS, region) {
          msg := fmt.Sprintf("Error: region %s is not a valid AWS region", region)
          s.ChannelMessageSend(
            m.ChannelID,
            msg,
          )
          return fmt.Errorf(msg), nil
        }
      }

      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf("Region: %s. Look good? (`>yes`/`>no`)", region),
      )
      config.Region = region
      return nil, nil
    },
  )

  askSize := NewState(
    "ask-size",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      slugPriceList := make([][]string, 0)
      slugSource := ""

      if (config.CloudProvider == "aws") {
        slugPriceList = AWS_VALID_SIZES
        slugSource = AWS_SIZE_SOURCE
      } else if (config.CloudProvider == "digitalocean") {
        slugPriceList = DIGITALOCEAN_VALID_SIZES
        slugSource = DIGITALOCEAN_SIZE_SOURCE
      } else {
        msg := "Cloud provider not set or invalid"
        s.ChannelMessageSend(
          m.ChannelID,
          msg,
        )
        return fmt.Errorf(msg), nil
      }

      msg := fmt.Sprintf(
        "Select the name of the size of the instance you'd like to use. To check the RAM+vCPU sizes and prices, you can go here: %s\nIf there's a size you don't see in the below list, you can request another here: %s\nTo set size, use `>size <size>`, e.g. `>size %s`",
        slugSource,
        ISSUE_LINK,
        slugPriceList[0][0],
      )

      sizeFields := []*discordgo.MessageEmbedField{}

      for _, slugPrice := range slugPriceList {
        slug := slugPrice[0]
        price := slugPrice[1]
        sizeFields = append(sizeFields, &discordgo.MessageEmbedField{
          Name: slug,
          Value: fmt.Sprintf("%s / hour", price),
          Inline: false,
        })
      }

      s.ChannelMessageSendComplex(
        m.ChannelID,
        &discordgo.MessageSend{
          Content: msg,
          Embed: &discordgo.MessageEmbed{
            Fields: sizeFields,
          },
        },
      )
      return nil, nil
    },
  )
  handleSize := NewState(
    "handle-size",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      size := args[1]
      slugPriceList := make([][]string, 0)
      if (config.CloudProvider == "aws") {
        slugPriceList = AWS_VALID_SIZES
      } else if (config.CloudProvider == "digitalocean") {
        slugPriceList = DIGITALOCEAN_VALID_SIZES
      } else {
        msg := "Cloud provider not set or invalid"
        s.ChannelMessageSend(
          m.ChannelID,
          msg,
        )
        return fmt.Errorf(msg), nil
      }

      for _, slugPrice := range slugPriceList {
        searchedSlug := slugPrice[0]
        if searchedSlug == size {
          s.ChannelMessageSend(
            m.ChannelID,
            fmt.Sprintf(
              "Selected Size: %s\nLook good? `>yes`/`>no`",
              size,
            ),
          )
          config.Size = size
          return nil, nil
        }
      }
      msg := fmt.Sprintf("Invalid size %s, pick a valid size ", size)
      s.ChannelMessageSend(
        m.ChannelID,
        msg,
      )
      return fmt.Errorf(msg), nil
    },
  )

  askSetupManagingRole := NewState(
    "ask-setup-managing-role",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      Guild, err := s.Guild(m.GuildID)
      possibleRoles := make(discordgo.Roles, 0)
      if err == nil {
        possibleRoles = Guild.Roles
      }
      fields := make([]*discordgo.MessageEmbedField, 0)
      for _, role := range possibleRoles {
        fields = append(fields, &discordgo.MessageEmbedField{
          Name: role.Name,
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
    },
  )
  handleSetupManagingRole := NewState(
    "handle-setup-managing-role",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
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
    },
  )

  askServerType := NewState(
    "ask-server-type",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf(
          "What type of server would you like to run? Note: You can only use one type of server at a time, but each server type will save its own world and state. \nIf you'd like to request a new mod or server type, make a feature request here: %s\nServer types: %s\n`>ephemeral set-type <type>`",
          ISSUE_LINK,
          strings.Join(SERVER_TYPES, ", "),
        ),
      )
      return nil, nil
    },
  )
  handleServerType := NewState(
    "handle-server-type",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      serverType := args[1]

      for _, validServerTypes := range SERVER_TYPES {
        if serverType == validServerTypes  {
          s.ChannelMessageSend(
            m.ChannelID,
            fmt.Sprintf("Server Type: %s\nConfirm? `>yes`/`>no`", serverType),
          )
          config.ServerType = serverType
          return nil, nil
        }
      }

      msg := "Invalid server type, make sure you've picked from the valid server type list."
      s.ChannelMessageSend(
        m.ChannelID,
        msg,
      )
      return fmt.Errorf(msg), nil
    },
  )

  saveConfig := NewState(
    "save-config",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      err := config.SaveConfig(kvConn)
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
      return nil, &rootState
    },
  )

  helpStep := NewState(
    "help",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      helpText := "EphemeralServer Help\n" +
      "`>ephemeral config`: Runs through the setup wizard. For first type setup or full reconfigurations.\n" +
      "`>ephemeral set-type`: Sets the type of server to boot up.\n" +
      "`>ephemeral set-size`: Sets the size of server to boot up. Use `>wq` when confirming to skip out of the rest of the wizard.\n"
      s.ChannelMessageSend(
        m.ChannelID,
        helpText,
      )
      return nil, &rootState
    },
  )

  ephemeralCtlState := NewState(
    "ephemeral-ctl-entrypoint",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) (error, *State) {
      actionString := args[1]
      var action EphemeralAction = NO_OP
      if actionString == "up" {
        action = CREATE
      }
      if action != NO_OP {
        RunEphemeral(ctx, action, config)
      }
      return nil, &rootState
    },
  )

  rootState.AddState(`^>eph[emeral]* config$`, configRoot)
  rootState.AddState(`^>eph[emeral]* set-type`, askServerType)
  rootState.AddState(`^>eph[emeral]* set-type (.+)`, handleServerType)
  rootState.AddState(`^>eph[emeral]* set-size`, askSize)
  rootState.AddState(`^>eph[emeral]* help$`, helpStep)

  rootState.AddState(`^>eph[emeral]* (up|down)$`, ephemeralCtlState)

  configRoot.AddState(`>cancel`, cancelStep)
  configRoot.AddState(`>continue`, askCloudProviderStep)

  askCloudProviderStep.AddState(`>cancel`, cancelStep)
  askCloudProviderStep.AddState(`>provider (digitalocean|aws)`, cloudProviderStep)

  cloudProviderStep.AddState(`>no`, askCloudProviderStep)
  cloudProviderStep.AddState(`>cancel`, cancelStep)
  cloudProviderStep.AddState(`>yes`, askRegion)

  askRegion.AddState(`>cancel`, cancelStep)
  askRegion.AddState(`>region (.+)`, handleRegion)

  handleRegion.AddState(`>no`, askRegion)
  handleRegion.AddState(`>cancel`, cancelStep)
  handleRegion.AddState(`>yes`, askCloudCredentials)

  askCloudCredentials.AddState(`>cancel`, cancelStep)
  askCloudCredentials.AddState(`>keys (.+) (.+)`, handleCloudCredentials)
  askCloudCredentials.AddState(`>keys (.+)`, handleCloudCredentials)

  handleCloudCredentials.AddState(`>no`, askCloudCredentials)
  handleCloudCredentials.AddState(`>cancel`, cancelStep)
  handleCloudCredentials.AddState(`>yes`, askSize)

  askSize.AddState(`>size (.*)`, handleSize)
  askSize.AddState(`>cancel`, cancelStep)

  handleSize.AddState(`>no`, askSize)
  handleSize.AddState(`>cancel`, cancelStep)
  handleSize.AddState(`>yes`, askSetupManagingRole)
  handleSize.AddState(`>wq`, saveConfig)

  askSetupManagingRole.AddState(`>skip`, askServerType)
  askSetupManagingRole.AddState(`>cancel`, cancelStep)
  askSetupManagingRole.AddState(`>role (.+)`, handleSetupManagingRole)

  handleSetupManagingRole.AddState(`>no`, askSetupManagingRole)
  handleSetupManagingRole.AddState(`>cancel`, cancelStep)
  handleSetupManagingRole.AddState(`>yes`, askServerType)

  askServerType.AddState(`>eph[emeral]* set-type (.+)`, handleServerType)
  askServerType.AddState(`>cancel`, cancelStep)

  handleServerType.AddState(`>yes`, saveConfig)
  handleServerType.AddState(`>cancel`, cancelStep)
  handleServerType.AddState(`>wq`, saveConfig)
  handleServerType.AddState(`>no`, askServerType)

  return rootState
}

