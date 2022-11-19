package internal

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

const (
	defaultBaseURL = "https://patrickscheid.de/s/"
	userAgent      = "secretli-cli"
)

type HTTPRemoteStore struct {
	client *resty.Client
}

func NewHTTPRemoteStore(baseURL string) *HTTPRemoteStore {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("User-Agent", userAgent)

	return &HTTPRemoteStore{client: client}
}

func (s *HTTPRemoteStore) Store(keySet KeySet, data EncryptedData, expiration string, burnAfterRead bool) error {
	type request struct {
		PublicID       string `json:"public_id"`
		RetrievalToken string `json:"retrieval_token"`
		DeletionToken  string `json:"deletion_token"`
		Nonce          string `json:"nonce"`
		EncryptedData  string `json:"encrypted_data"`
		Expiration     string `json:"expiration"`
		BurnAfterRead  bool   `json:"burn_after_read"`
	}

	dto := request{
		PublicID:       keySet.PublicID,
		RetrievalToken: keySet.RetrievalToken,
		DeletionToken:  keySet.DeletionToken,
		Nonce:          B64Encode(data.Nonce),
		EncryptedData:  B64Encode(data.Cipher),
		Expiration:     expiration,
		BurnAfterRead:  burnAfterRead,
	}

	resp, err := s.client.R().
		SetBody(dto).
		Post("api/secret")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("error sharing secret: %d", resp.StatusCode())
	}

	return nil
}

func (s *HTTPRemoteStore) Load(keySet KeySet) (EncryptedData, error) {
	type response struct {
		Nonce         string `json:"nonce"`
		EncryptedData string `json:"encrypted_data"`
	}

	var dto response

	resp, err := s.client.R().
		SetPathParam("id", keySet.PublicID).
		SetHeader("X-Retrieval-Token", keySet.RetrievalToken).
		SetResult(&dto).
		Post("api/secret/{id}")

	if err != nil {
		return EncryptedData{}, err
	}

	if resp.IsError() {
		return EncryptedData{}, fmt.Errorf("error loading secret: %d", resp.StatusCode())
	}

	nonce, err := B64Decode(dto.Nonce)
	if err != nil {
		return EncryptedData{}, err
	}

	cipher, err := B64Decode(dto.EncryptedData)
	if err != nil {
		return EncryptedData{}, err
	}

	data := EncryptedData{
		Nonce:  nonce,
		Cipher: cipher,
	}
	return data, nil
}

func (s *HTTPRemoteStore) Delete(keySet KeySet, deletionToken string) error {
	resp, err := s.client.R().
		SetPathParam("id", keySet.PublicID).
		SetHeader("X-Retrieval-Token", keySet.RetrievalToken).
		SetHeader("X-Deletion-Token", deletionToken).
		Delete("api/secret/{id}")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("error deleting secret: %d", resp.StatusCode())
	}

	return nil
}
