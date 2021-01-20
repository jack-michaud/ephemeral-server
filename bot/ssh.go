package bot

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

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
