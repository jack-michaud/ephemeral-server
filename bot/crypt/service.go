package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
)

type CryptService struct {
  gcm cipher.AEAD
}

func NewCryptService(key []byte) CryptService {
  block, err := aes.NewCipher(key)
  if err != nil {
    log.Fatalf("error creating cipher: %s", err)
  }
  gcm, err := cipher.NewGCM(block)
  if err != nil {
    log.Fatalf("error creating gcm: %s", err)
  }
  return CryptService{
    gcm,
  }
}

func (s CryptService) Encrypt(plaintext []byte) ([]byte, error)  {
  nonce := s.GenerateNonce()
  encrypted := s.gcm.Seal(nil, nonce, plaintext, nil)
  return append(nonce, encrypted...), nil
}

func (s CryptService) Decrypt(encrypted []byte) ([]byte, error)  {
  nonce, encrypted := encrypted[:s.gcm.NonceSize()], encrypted[s.gcm.NonceSize():]
  decrypted, err := s.gcm.Open(nil, nonce, encrypted, nil)
  if err != nil {
    return nil, fmt.Errorf("Error decrypting: %s", err)
  }
  return decrypted, nil
}

func (s CryptService) GenerateNonce() []byte {
  nonce := make([]byte, s.gcm.NonceSize())
  if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
    log.Fatalln("couldnt make random numbers:", err)
  }
  return nonce
}


