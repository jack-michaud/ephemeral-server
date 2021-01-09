package bot

import (
	"sync"
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
  ServerType string
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

