package store

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)

type KVConsul struct {
	kv     *api.KV
	client *api.Client
}

func (c KVConsul) Get(Id string) ([]byte, error) {
	kvPair, _, err := c.kv.Get(Id, nil)
	if kvPair == nil {
		return nil, fmt.Errorf("could not find key %s", Id)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get key value: %s", err)
	}
	return kvPair.Value, nil
}

func (c KVConsul) Set(Id string, value []byte) error {
	_, err := c.kv.Put(&api.KVPair{
		Key:   Id,
		Value: value,
	}, nil)
	if err != nil {
		return fmt.Errorf("could not get key value: %s", err)
	}
	return nil
}

func (c KVConsul) Cleanup() error {
	// nothing to do
	return nil
}

func (c KVConsul) TestLive() error {
	_, err := c.client.Status().Leader()
	return err
}

type KVConsulConfig = api.Config

func NewKVConsul(config *KVConsulConfig) (IKVStore, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("could not create consul client: %s", err)
	}

	kv := client.KV()
	var conn IKVStore = KVConsul{
		kv:     kv,
		client: client,
	}
	return conn, nil
}
