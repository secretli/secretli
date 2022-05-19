package internal

import (
	"net/http"
)

type HTTPRemoteStore struct {
	client *Client
}

func NewHTTPRemoteStore(client *Client) *HTTPRemoteStore {
	return &HTTPRemoteStore{client: client}
}

func (s *HTTPRemoteStore) Store(keySet KeySet, data EncryptedData, expiration string) error {
	type request struct {
		PublicID       string `json:"public_id"`
		RetrievalToken string `json:"retrieval_token"`
		Nonce          string `json:"nonce"`
		EncryptedData  string `json:"encrypted_data"`
		Expiration     string `json:"expiration"`
	}

	dto := request{
		PublicID:       keySet.PublicID(),
		RetrievalToken: keySet.RetrievalToken(),
		Nonce:          data.Nonce,
		EncryptedData:  data.Cipher,
		Expiration:     expiration,
	}

	req, err := s.client.NewRequest(http.MethodPost, "api/secret", dto)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

func (s *HTTPRemoteStore) Load(keySet KeySet) (EncryptedData, error) {
	type response struct {
		Nonce         string `json:"nonce"`
		EncryptedData string `json:"encrypted_data"`
	}

	req, err := s.client.NewRequest(http.MethodPost, "api/secret/"+keySet.PublicID(), nil)
	if err != nil {
		return EncryptedData{}, err
	}
	req.Header.Set("X-Retrieval-Token", keySet.RetrievalToken())

	var dto response
	_, err = s.client.Do(req, &dto)
	if err != nil {
		return EncryptedData{}, err
	}

	return EncryptedData{
		Nonce:  dto.Nonce,
		Cipher: dto.EncryptedData,
	}, nil
}
