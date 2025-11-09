package redditclient

type MemoryTokenStore struct {
	hasToken bool
	token    SavedToken
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{}
}

func (store *MemoryTokenStore) SaveToken(token SavedToken) error {
	store.token = token
	store.hasToken = true
	return nil
}

func (store *MemoryTokenStore) GetToken() (SavedToken, bool, error) {
	if store.hasToken {
		return store.token, true, nil
	}
	return store.token, false, nil
}
