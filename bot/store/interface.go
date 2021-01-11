package store

type IKVStore interface {
  TestLive() error

  // result will be null if does not exist
  Get(Id string) ([]byte, error)
  Set(Id string, value []byte) error

  Cleanup() error
}
