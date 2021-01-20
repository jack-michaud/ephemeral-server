package bot_tests

import (
	"testing"

	"github.com/jack-michaud/ephemeral-server/bot"
)

func TestSshGen(t *testing.T) {
	privateKey, err := bot.GeneratePrivateKey()
	if err != nil {
		t.Errorf("failed to generate private key: %s", err)
		t.FailNow()
	}

	publicKeyString, err := bot.GetAuthorizedFilePublicKeyString(privateKey)
	t.Log(string(publicKeyString))
	if err != nil {
		t.Errorf("failed to generate publickey string key: %s", err)
		t.FailNow()
	}

	privateKeyString := bot.GetPrivateKeyString(privateKey)
	t.Log(string(privateKeyString))
}
