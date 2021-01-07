package bot

import (
  "fmt"
	"log"
	"regexp"
	"sync"
	"github.com/bwmarrin/discordgo"
)

type AwsCreds struct {
}
type DigitalOceanCreds struct {
  AccessKey string
}

type PrivateKey = []byte

type Config struct {
  CloudProvider string
  DigitalOceanCreds *DigitalOceanCreds
  // Aws.
  AwsCreds *AwsCreds
  // Private key to access the VPS.
  PrivateKey *PrivateKey
  ManagingRoleId string
}

func GetConfigForServerId(Id string) (Config, error) {
  return Config{
    CloudProvider: "digitalocean",
    DigitalOceanCreds: &DigitalOceanCreds{
      AccessKey: "123",
    },
    AwsCreds: nil,
    PrivateKey: nil,
  }, nil
}

func (c *Config) SaveConfig() {
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

    strings := rg.FindAllString(messageContent, -1)
    if strings != nil {
      log.Println("Found matching state:", nextState.id)
      err := nextState.Action(session, message, config, strings)
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

  firstStep := NewState(
    "first-step",
    func(s *discordgo.Session, m *discordgo.MessageCreate, config *Config, args []string) error {
      s.ChannelMessageSend(
        m.ChannelID,
        fmt.Sprintf("Argument: %s", args[0]),
      )
      return nil
    },
  )

  configRoot.AddState(`>cancel`, cancelStep)
  configRoot.AddState(`>continue (.*)`, firstStep)

  rootState := NewState("root", NilAction)
  rootState.AddState(`^>ephemeral config.*`, configRoot)
  cancelStep.AddState(`^>ephemeral config.*`, configRoot)

  return rootState
}

