package serverbridge

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/jack-michaud/ephemeral-server/bot/store"
	"golang.org/x/crypto/ssh"
)

type PrivateKey = rsa.PrivateKey

type ConnectOptions struct {
	ServerIpAddress *string
	PrivateKey      *PrivateKey
}

func ConnectToServer(ctx context.Context, options *ConnectOptions, kvConn store.IKVStore) (*ssh.Client, error) {
	if options.ServerIpAddress == nil {
		return nil, fmt.Errorf("Could not find stored server IP address. Is the server up?")
	}
	serverIpAddressString := *options.ServerIpAddress

	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(options.PrivateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(&block)
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)

	if err != nil {
		return nil, fmt.Errorf("could not parse authorized key: %s", err)
	}
	log.Println("Trying to connect to", serverIpAddressString)
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", serverIpAddressString), &ssh.ClientConfig{
		User: "minecraft",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	})

	if err != nil {
		return nil, fmt.Errorf("could not connect to ssh server: %s", err)
	}

	return client, nil
}
