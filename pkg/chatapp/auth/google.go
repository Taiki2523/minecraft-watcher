package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GoogleClient struct {
	Config *oauth2.Config
}

type UserInfo struct {
	GoogleID  string
	Email     string
	Name      string
	AvatarURL string
}

func NewGoogleClient(clientID, clientSecret, redirectURL string) *GoogleClient {
	return &GoogleClient{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				oauth2.UserinfoEmailScope,
				oauth2.UserinfoProfileScope,
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (gc *GoogleClient) AuthCodeURL(state string) string {
	return gc.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (gc *GoogleClient) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return gc.Config.Exchange(ctx, code)
}

func (gc *GoogleClient) FetchUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	service, err := oauth2.NewService(ctx, option.WithTokenSource(gc.Config.TokenSource(ctx, token)))
	if err != nil {
		return UserInfo{}, fmt.Errorf("create oauth2 service: %w", err)
	}

	info, err := service.Userinfo.Get().Do()
	if err != nil {
		return UserInfo{}, fmt.Errorf("get userinfo: %w", err)
	}

	return UserInfo{
		GoogleID:  info.Id,
		Email:     info.Email,
		Name:      info.Name,
		AvatarURL: info.Picture,
	}, nil
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func RedirectWithError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
