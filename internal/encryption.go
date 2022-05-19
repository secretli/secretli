package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

type EncryptedData struct {
	Nonce  string
	Cipher string
}

type KeySet struct {
	shareSecret    []byte
	shareKey       []byte
	publicID       []byte
	retrievalToken []byte
	deletionToken  []byte
	cipher         cipher.AEAD
}

func NewRandomKeySet() (KeySet, error) {
	shareSecret, err := generateRandomBytes(32)
	if err != nil {
		return KeySet{}, err
	}
	return deriveKeysFromBytes(shareSecret)
}

func NewRandomKeySetWithPassword(password string) (KeySet, error) {
	shareSecret, err := generateRandomBytes(32)
	if err != nil {
		return KeySet{}, err
	}

	masterKey := pbkdf2.Key([]byte(password), shareSecret, 100000, 32, sha512.New)
	keySet, err := deriveKeysFromBytes(masterKey)
	if err != nil {
		return KeySet{}, err
	}

	keySet.shareSecret = shareSecret
	return keySet, nil
}

func KeySetFromString(key string) (KeySet, error) {
	shareSecret, err := decode(key)
	if err != nil {
		return KeySet{}, err
	}
	return deriveKeysFromBytes(shareSecret)
}

func KeySetWithPasswordFromString(key string, password string) (KeySet, error) {
	shareSecret, err := decode(key)
	if err != nil {
		return KeySet{}, err
	}

	masterKey := pbkdf2.Key([]byte(password), shareSecret, 100000, 32, sha512.New)
	keySet, err := deriveKeysFromBytes(masterKey)
	if err != nil {
		return KeySet{}, err
	}

	keySet.shareSecret = shareSecret
	return keySet, nil
}

func (k KeySet) ShareSecret() string {
	return encode(k.shareSecret)
}

func (k KeySet) PublicID() string {
	return encode(k.publicID)
}

func (k KeySet) RetrievalToken() string {
	return encode(k.retrievalToken)
}

func (k KeySet) DeletionToken() string {
	return encode(k.deletionToken)
}

func (k KeySet) Encrypt(plaintext string) (data EncryptedData, err error) {
	nonce, err := generateRandomBytes(k.cipher.NonceSize())
	if err != nil {
		return
	}

	plainBytes := []byte(plaintext)
	ciphertext := k.cipher.Seal(nil, nonce, plainBytes, nil)

	data.Nonce = encode(nonce)
	data.Cipher = encode(ciphertext)
	return
}

func (k KeySet) Decrypt(data EncryptedData) (string, error) {
	nonce, err1 := decode(data.Nonce)
	ciphertext, err2 := decode(data.Cipher)

	err := multierror.Append(err1, err2).ErrorOrNil()
	if err != nil {
		return "", err
	}

	plaintext, err := k.cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func generateRandomBytes(n int) ([]byte, error) {
	result := make([]byte, n)
	_, err := rand.Read(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func deriveKeysFromBytes(key []byte) (KeySet, error) {
	shareKey, err1 := deriveSubkey(key, "share_item_encryption_key", 32)
	publicID, err2 := deriveSubkey(key, "share_item_uuid", 16)
	retrievalToken, err3 := deriveSubkey(key, "share_item_token", 16)
	deletionToken, err4 := generateRandomBytes(16)

	err := multierror.Append(err1, err2, err3, err4).ErrorOrNil()
	if err != nil {
		return KeySet{}, err
	}

	gcm, err := setupCipher(shareKey)
	if err != nil {
		return KeySet{}, err
	}

	return KeySet{
		shareSecret:    key,
		shareKey:       shareKey,
		publicID:       publicID,
		retrievalToken: retrievalToken,
		deletionToken:  deletionToken,
		cipher:         gcm,
	}, nil
}

func deriveSubkey(key []byte, info string, length int) ([]byte, error) {
	reader := hkdf.Expand(sha512.New, key, []byte(info))
	result := make([]byte, length)
	_, err := io.ReadFull(reader, result)
	return result, err
}

func setupCipher(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func encode(input []byte) string {
	return base64.RawURLEncoding.EncodeToString(input)
}

func decode(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}
