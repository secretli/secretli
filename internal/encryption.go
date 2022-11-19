package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

type EncryptedData struct {
	Nonce  []byte
	Cipher []byte
}

type KeySet struct {
	ShareSecret    string
	ShareKey       string
	PublicID       string
	RetrievalToken string
	DeletionToken  string
	cipher         cipher.AEAD
}

func NewKeySet(password string) (KeySet, error) {
	shareSecret, err := generateRandomBytes(32)
	if err != nil {
		return KeySet{}, err
	}

	if password == "" {
		return deriveKeysFromBytes(shareSecret)
	}

	masterKey := pbkdf2.Key([]byte(password), shareSecret, 100_000, 32, sha512.New)
	keySet, err := deriveKeysFromBytes(masterKey)
	if err != nil {
		return KeySet{}, err
	}

	keySet.ShareSecret = B64Encode(shareSecret)
	return keySet, nil
}

func KeySetFromString(key string, password string) (KeySet, error) {
	shareSecret, err := B64Decode(key)
	if err != nil {
		return KeySet{}, err
	}

	if password == "" {
		return deriveKeysFromBytes(shareSecret)
	}

	masterKey := pbkdf2.Key([]byte(password), shareSecret, 100_000, 32, sha512.New)
	keySet, err := deriveKeysFromBytes(masterKey)
	if err != nil {
		return KeySet{}, err
	}

	keySet.ShareSecret = B64Encode(shareSecret)
	return keySet, nil
}

func (k KeySet) Encrypt(plaintext string) (EncryptedData, error) {
	nonce, err := generateRandomBytes(k.cipher.NonceSize())
	if err != nil {
		return EncryptedData{}, err
	}

	plainBytes := []byte(plaintext)
	ciphertext := k.cipher.Seal(nil, nonce, plainBytes, nil)

	data := EncryptedData{
		Nonce:  nonce,
		Cipher: ciphertext,
	}

	return data, nil
}

func (k KeySet) Decrypt(data EncryptedData) (string, error) {
	plaintext, err := k.cipher.Open(nil, data.Nonce, data.Cipher, nil)
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

	keySet := KeySet{
		ShareSecret:    B64Encode(key),
		ShareKey:       B64Encode(shareKey),
		PublicID:       B64Encode(publicID),
		RetrievalToken: B64Encode(retrievalToken),
		DeletionToken:  B64Encode(deletionToken),
		cipher:         gcm,
	}

	return keySet, nil
}

func deriveSubkey(key []byte, info string, length int) ([]byte, error) {
	reader := hkdf.New(sha512.New, key, nil, []byte(info))
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
