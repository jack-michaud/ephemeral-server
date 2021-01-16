package store

import "fmt"

type MockKVStore struct {
  internalMap map[string] []byte
}

func NewMockKVStore() IKVStore {
  return MockKVStore{
    internalMap: make(map[string] []byte),
  }
}

func (kv MockKVStore) TestLive() error {
  return nil
}

func (kv MockKVStore) Get(Id string) ([]byte, error) {
  if kv.internalMap[Id] == nil {
    return nil, fmt.Errorf("could not find given id")
  }
  value := kv.internalMap[Id]
  return value, nil
}

func (kv MockKVStore) Set(Id string, value []byte) error {
  kv.internalMap[Id] = value
  return nil
}

func (kv MockKVStore) Cleanup() error {
  return nil
}

