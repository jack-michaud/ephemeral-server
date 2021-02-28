package bot

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/jack-michaud/ephemeral-server/bot/serverbridge"
	"github.com/jack-michaud/ephemeral-server/bot/store"
	"golang.org/x/crypto/ssh"
	//"golang.org/x/crypto/ssh"
)

type PrivateKey = rsa.PrivateKey

// I'd like to use elliptic curve keys, but that's pending this issue:
// https://github.com/golang/go/issues/33564
func GeneratePrivateKey() (*PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return nil, fmt.Errorf("error generating private Key: %s", err)
	}

	return privateKey, nil
}

func GetAuthorizedFilePublicKeyString(pk *PrivateKey) ([]byte, error) {
	publicKey, err := ssh.NewPublicKey(pk.Public())
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(publicKey), nil
}

func GetPrivateKeyString(pk *PrivateKey) []byte {
	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	}
	bytes := pem.EncodeToMemory(&block)
	return bytes
}

// Opens an ssh connection to a given server ID's VPS.
// If a server isn't created (no IP stored for it) an error will be returned.
// If no private key is stored, an error will be returned.
func ConnectToServerFromServerId(ctx context.Context, Id string, conn store.IKVStore) (*ssh.Client, error) {
	config, err := GetConfigForServerId(Id, conn)
	if err != nil {
		return nil, fmt.Errorf("could not get config: %s", err)
	}

	if config.ServerIpAddress == nil {
		return nil, fmt.Errorf("Could not connect to server: No IP. Is the server up?")
	}
	if config.PrivateKey == nil {
		return nil, fmt.Errorf(
			"Could not connect to server: No private key. Have you started the server before?",
		)
	}

	sshClient, err := serverbridge.ConnectToServer(ctx, &serverbridge.ConnectOptions{
		ServerIpAddress: config.ServerIpAddress,
		PrivateKey:      config.PrivateKey,
	}, conn)

	if err != nil {
		return nil, fmt.Errorf("could not connect to server: %s", err)
	}

	return sshClient, nil
}
