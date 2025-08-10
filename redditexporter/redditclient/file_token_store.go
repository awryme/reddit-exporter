package redditclient

import (
	"errors"
	"fmt"

	"github.com/awryme/reddit-exporter/pkg/jsonfile"
)

type FileTokenStore struct {
	tokenfile string
}

func NewFileTokenStore(tokenfile string) *FileTokenStore {
	return &FileTokenStore{tokenfile}
}

func (store *FileTokenStore) SaveToken(token SavedToken) error {
	return jsonfile.Write(store.tokenfile, token)
}

func (store *FileTokenStore) GetToken() (SavedToken, bool, error) {
	token, err := jsonfile.Read[SavedToken](store.tokenfile)
	if errors.Is(err, jsonfile.ErrFileNotFound) {
		return token, false, nil
	}
	if err != nil {
		return token, false, fmt.Errorf("read token from file: %w", err)
	}
	return token, true, nil
}
