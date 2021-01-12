package crypt

import (
	"encoding/hex"
	"testing"
)

func TestEncDec(t *testing.T) {

  key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
  s := NewCryptService(key)

  plaintext := "   test string blah blah blah    "
  encrypted, err := s.Encrypt([]byte(plaintext))
  if err != nil {
    t.Error("error encrypting:", err)
  }
  decrypted, err := s.Decrypt(encrypted)
  if err != nil {
    t.Error("error decrypting:", err)
  }
  if string(plaintext) != string(decrypted) {
    t.Errorf("bad decrypt: %s != %s", plaintext, decrypted)
  }
}

func TestTamperProof(t *testing.T) {

  key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
  s := NewCryptService(key)

  plaintext := "   test string blah blah blah    "
  encrypted, err := s.Encrypt([]byte(plaintext))
  if err != nil {
    t.Error("error encrypting:", err)
  }
  encrypted[24] = byte(4)
  _, err = s.Decrypt(encrypted)
  if err == nil {
    t.Error("Expected error after tampering with data, did not get any")
  }
}
