package vendors

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type CryptoRCAKeyOptions struct {
	Size int
}

type CryptoRCAEncryptOptions struct {
	Text      string
	PublicKey string
}

type CryptoRCADecryptOptions struct {
	Text       string
	PrivateKey string
}

type Crypto struct {
}

func (c *Crypto) CustomRCAGenerateKey(keyOptions CryptoRCAKeyOptions) ([]byte, error) {

	key, err := rsa.GenerateKey(rand.Reader, keyOptions.Size)
	if err != nil {
		return nil, err
	}

	privatePem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	privateBytes := pem.EncodeToMemory(privatePem)

	publicPem := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}
	publicBytes := pem.EncodeToMemory(publicPem)

	arr := [][]byte{privateBytes, publicBytes}

	return bytes.Join(arr, []byte("")), nil
}

func (c *Crypto) RCAGenerateKey() ([]byte, error) {
	return c.CustomRCAGenerateKey(CryptoRCAKeyOptions{Size: 2048})
}

func (c *Crypto) CustomRCAEncrypt(encryptOptions CryptoRCAEncryptOptions) ([]byte, error) {

	block, _ := pem.Decode([]byte(encryptOptions.PublicKey))
	if block == nil {
		return nil, errors.New("public key is invalid")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	text, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(encryptOptions.Text), nil)
	if err != nil {
		return nil, err
	}

	return []byte(text), nil
}

func (c *Crypto) CustomRCADecrypt(decryptOptions CryptoRCADecryptOptions) ([]byte, error) {

	block, _ := pem.Decode([]byte(decryptOptions.PrivateKey))
	if block == nil {
		return nil, errors.New("private key is invalid")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	text, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, []byte(decryptOptions.Text), nil)
	if err != nil {
		return nil, err
	}

	return []byte(text), nil
}

func NewCrypto() *Crypto {
	return &Crypto{}
}
