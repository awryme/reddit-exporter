package redditclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/awryme/reddit-exporter/pkg/xhttp"
	"github.com/awryme/slogf"
	"github.com/denisbrodbeck/machineid"
)

const (
	redditGrantType = "https://oauth.reddit.com/grants/installed_client"
	redditAuthUrl   = "https://www.reddit.com/api/v1/access_token"
	redditUserAgent = "reddit-exporter/v1.1"
)

const tokenExpBefore = time.Hour

type authResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   uint64 `json:"expires_in"`
}

type SavedToken struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

const machineidAppID = "reddit-exporter"

type TokenStore interface {
	SaveToken(token SavedToken) error
	GetToken() (SavedToken, bool, error)
}

type AuthService struct {
	clientID     string
	clientSecret string
	tokenStore   TokenStore

	logf       slogf.Logf
	httpClient *http.Client
}

func NewAuth(log slog.Handler, clientID string, clientSecret string, tokenStore TokenStore) *AuthService {
	httpClient := xhttp.NewClient()
	return &AuthService{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenStore:   tokenStore,
		logf:         slogf.New(log),
		httpClient:   httpClient,
	}
}

func (svc *AuthService) Auth() (string, error) {
	token, ok, err := svc.tokenStore.GetToken()
	if err != nil {
		return "", fmt.Errorf("retrieve token from store: %w", err)
	}

	if !ok {
		svc.logf("authenticating: no token in store")
		return svc.authAndStoreFile()
	}

	if time.Now().Add(tokenExpBefore).After(token.Expires) {
		svc.logf("authenticating: token expires soon", slog.Time("expires in", token.Expires))
		return svc.authAndStoreFile()
	}
	return token.Token, nil
}

func (svc *AuthService) ForceAuth() (string, error) {
	return svc.authAndStoreFile()
}

func (svc *AuthService) authAndStoreFile() (string, error) {
	res, err := svc.httpAuth()
	if err != nil {
		return "", fmt.Errorf("http auth: %w", err)
	}

	savedToken := SavedToken{
		Token:   res.AccessToken,
		Expires: time.Now().Add(time.Second * time.Duration(res.ExpiresIn)),
	}

	err = svc.tokenStore.SaveToken(savedToken)
	if err != nil {
		return "", fmt.Errorf("save token to store: %w", err)
	}
	return savedToken.Token, nil
}

func (svc *AuthService) httpAuth() (*authResponse, error) {
	if svc.clientID == "" || svc.clientSecret == "" {
		return nil, fmt.Errorf("client_id or client_secret is empty")
	}
	deviceID, err := getDeviceID()
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("grant_type", redditGrantType)
	form.Set("device_id", deviceID)

	authReq, _ := http.NewRequest(http.MethodPost, redditAuthUrl, strings.NewReader(form.Encode()))
	authReq.SetBasicAuth(svc.clientID, svc.clientSecret)
	authReq.Header.Set("User-Agent", redditUserAgent)

	resp, err := svc.httpClient.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("make auth http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bad status body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read bad status body: %w", err)
		}
		return nil, fmt.Errorf("auth http bad status %d: %w, body: '%s'", resp.StatusCode, err, string(body))
	}

	var tokenResp authResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("decode json auth resp: %w, body: '%s'", err, string(body))
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("decode json auth resp: token is empty, body: '%s'", string(body))
	}

	return &tokenResp, nil
}

func getDeviceID() (string, error) {
	// reddit required device_id to be 20-30 characters
	const minLen = 20
	const maxLen = 30

	id, err := machineid.ProtectedID(machineidAppID)
	if err != nil {
		return "", fmt.Errorf("generate machine id: %w", err)
	}

	for len(id) < minLen {
		// duplicate until 20 chars
		id += id
	}
	if len(id) > maxLen {
		return id[:maxLen], nil
	}
	return id, nil
}
