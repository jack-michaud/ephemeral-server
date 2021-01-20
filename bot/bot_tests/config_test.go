package bot_tests

import (
	"github.com/jack-michaud/ephemeral-server/bot"
	"github.com/jack-michaud/ephemeral-server/bot/store"
	"testing"
)

func TestConfigPrivateKey(t *testing.T) {
	kvStore := store.NewMockKVStore()

	ServerId := "testId-123"
	config, _ := bot.GetConfigForServerId(ServerId, kvStore)

	if config.PrivateKey != nil {
		t.Error("Private key is not none")
		t.FailNow()
	}

	PrivateKey, err := bot.GeneratePrivateKey()
	if err != nil {
		t.Errorf("Got error generating private key: %s", err)
		t.FailNow()
	}

	config.PrivateKey = PrivateKey
	config.SaveConfig(kvStore)

	// Test retrieval
	config, err = bot.GetConfigForServerId(ServerId, kvStore)
	if err != nil {
		t.Errorf("Got error fetching config: %s", err)
		t.FailNow()
	}
	if config.PrivateKey == nil {
		t.Error("Could not find privatekey from config")
		t.FailNow()
	}

	if !config.PrivateKey.Equal(PrivateKey) {
		t.Error("Retrieved private key is not the same as generated one")
		t.FailNow()
	}
}
