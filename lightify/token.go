package lightify

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"
)

var (
	lightifyToken = ""

	LightifyConfig oauth2.Config
)

func NewConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     os.Getenv("LIGHTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("LIGHTIFY_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("LIGHTIFY_REDIRECT_URL"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://emea.lightify-api.com/oauth2/authorize",
			TokenURL: "https://emea.lightify-api.com/oauth2/access_token",
		},
	}
}

func GetToken() *oauth2.Token {
	return &oauth2.Token{AccessToken: lightifyToken}
}

func GenerateToken() (string, error) {
	refreshToken, err := ioutil.ReadFile("token.txt")
	if err != nil {
		return "", err
	}

	response, err := http.PostForm("https://emea.lightify-api.com/oauth2/access_token", url.Values{
		"client_id":     {LightifyConfig.ClientID},
		"client_secret": {LightifyConfig.ClientSecret},
		"refresh_token": {string(refreshToken)},
		"grant_type":    {"refresh_token"},
	})
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	var newTokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err = json.NewDecoder(response.Body).Decode(&newTokens); err != nil {
		return "", err
	}

	if err = ioutil.WriteFile("token.txt", []byte(newTokens.RefreshToken), 0755); err != nil {
		return "", err
	}

	return newTokens.AccessToken, nil
}

func RefreshTokenRoutine() {
	for {
		time.Sleep(time.Hour * 48)

		token, err := GenerateToken()
		if err != nil {
			log.Println(err)
			continue
		}

		lightifyToken = token
	}
}
