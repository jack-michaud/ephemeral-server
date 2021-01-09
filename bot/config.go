package bot

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
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
  ServerId string
  CloudProvider string
  DigitalOceanCreds *DigitalOceanCreds
  Region string
  Size string
  // Aws.
  AwsCreds *AwsCreds
  // Private key to access the VPS.
  PrivateKey *PrivateKey
  ManagingRoleId string
  ServerType string
}

// Gets existing config from store or creates new config for server
func GetConfigForServerId(Id string, conn *redis.Client) (*Config, error) {
  ret := conn.Exists(Id)
  if ret.Val() == 1 {
    stringRet := conn.Get(Id)
    b, err := stringRet.Bytes()
    if err != nil {
      return nil, fmt.Errorf("error getting config from store: %s", err)
    }
    config := &Config{}
    err = json.Unmarshal(b, config)
    if err != nil {
      return nil, fmt.Errorf("error deserializing config: %s", err)
    }
    return config, nil
  } else {
    // Default config
    return &Config{
      ServerId: Id,
    }, nil
  }
}

func (c *Config) SaveConfig(conn *redis.Client) error {
  data, err := json.Marshal(*c)
  if err != nil {
    return fmt.Errorf("could not serialize config: %s", err)
  }
  ret := conn.Set(c.ServerId, data, time.Duration(0))
  retVal := ret.Val()
  if retVal == "OK" {
    return nil
  } else {
    return fmt.Errorf("could not save config: got error code from store: %s", retVal)
  }
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

