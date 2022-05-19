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
		PublicID:       keySet.PublicID(),
		RetrievalToken: keySet.RetrievalToken(),
		DeletionToken:  keySet.DeletionToken(),
		Nonce:          data.Nonce,
		EncryptedData:  data.Cipher,
		Expiration:     expiration,
		BurnAfterRead:  burnAfterRead,
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

func (s *HTTPRemoteStore) Delete(keySet KeySet, deletionToken string) error {
	req, err := s.client.NewRequest(http.MethodDelete, "api/secret/"+keySet.PublicID(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Retrieval-Token", keySet.RetrievalToken())
	req.Header.Set("X-Deletion-Token", deletionToken)

	_, err = s.client.Do(req, nil)
	if err != nil {
		return err
	}
	return nil
}
