package utils

import (
	"context"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleInfo struct {
	Provide_id string `json:"sub"`
	Fullname   string `json:"name"`
	Email      string `json:"email"`
	Picture    string `json:"picture"`
}

func GetGoogleAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "1015595862932-17dkqj34ae6leooup24kgfro5on96bod.apps.googleusercontent.com",
		ClientSecret: "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",
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

}

func GetGoogleLoginURL(config *oauth2.Config, state string) string {
	return config.AuthCodeURL(state)
}

func GetGoogleUserInfo(config *oauth2.Config, code string) (string, error) {
	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return "", fmt.Errorf("Failed to exchange token: %v", err)
	}
	client := config.Client(context.Background(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", fmt.Errorf("Failed to get user info: %v", err)
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response: %v", err)
	}

	return string(data), nil
}
