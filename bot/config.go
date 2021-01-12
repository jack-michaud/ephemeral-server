package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/jack-michaud/ephemeral-server/bot/crypt"
	"github.com/jack-michaud/ephemeral-server/bot/store"
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

func GetSecretKey() []byte {
  return []byte(os.Getenv("SECRET_KEY"))
}

// Gets existing config from store or creates new config for server
func GetConfigForServerId(Id string, conn store.IKVStore) (*Config, error) {
  rawConfig, err := conn.Get(fmt.Sprintf("configs/%s", Id))
  if rawConfig != nil {
    if err != nil {
      return nil, fmt.Errorf("error getting config from store: %s", err)
    }

    svc := crypt.NewCryptService(GetSecretKey())
    decryptedBytes, err := svc.Decrypt(rawConfig)
    if err != nil {
      return nil, fmt.Errorf("decryption error: %s", err)
    }

    config := &Config{}
    err = json.Unmarshal(decryptedBytes, config)
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

func (c *Config) SaveConfig(conn store.IKVStore) error {
  data, err := json.Marshal(*c)
  if err != nil {
    return fmt.Errorf("could not serialize config: %s", err)
  }
  svc := crypt.NewCryptService(GetSecretKey())
  encryptedBytes, err := svc.Encrypt(data)
  if err != nil {
    return fmt.Errorf("encryption error: %s", err)
  }
  err = conn.Set(fmt.Sprintf("configs/%s", c.ServerId), encryptedBytes)
  if err == nil {
    return nil
  } else {
    return fmt.Errorf("could not save config: got error code from store: %s", err)
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

