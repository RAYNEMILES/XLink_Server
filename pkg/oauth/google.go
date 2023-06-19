package oauth

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/people/v1"
	"strings"
	"time"
)

func ParseGoogleCode(platformId int, code, idToken string) (map[string]interface{}, error) {
	if idToken != "" {
		return parseGoogleIdToken(platformId, idToken)
	} else {
		return parseGoogleCode(platformId, code)
	}
}

func parseGoogleCode(platformId int, code string) (map[string]interface{}, error) {
	conf := &oauth2.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		RedirectURL:  "YOUR_REDIRECT_URL",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	switch platformId {
	case constant.WebPlatformID:
		conf.ClientID = config.Config.Oauth.Google.Web.ClientID
		conf.ClientSecret = config.Config.Oauth.Google.Web.ClientSecret
		break
	case constant.MiniWebPlatformID:
		conf.ClientID = config.Config.Oauth.Google.Web.ClientID
		conf.ClientSecret = config.Config.Oauth.Google.Web.ClientSecret
		break
	case constant.IOSPlatformID:
		conf.ClientID = config.Config.Oauth.Google.Ios.ClientID
		conf.ClientSecret = config.Config.Oauth.Google.Ios.ClientSecret
		break
	case constant.AndroidPlatformID:
		conf.ClientID = config.Config.Oauth.Google.Android.ClientID
		conf.ClientSecret = config.Config.Oauth.Google.Android.ClientSecret
		break
	}

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	client := conf.Client(context.Background(), token)
	service, err := people.New(client)
	if err != nil {
		return nil, err
	}

	res, err := service.People.Get("people/me").Do()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":    res.ResourceName,
		"email": res.EmailAddresses[0].Value,
		"name":  res.Names[0].DisplayName,
	}, nil
}

func parseGoogleIdToken(platformId int, idToken string) (map[string]interface{}, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID token: %s", idToken)
	}

	payload, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid ID token: %s", idToken)
	}

	// Unmarshal the payload into a map.
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("invalid ID token: %s", idToken)
	}

	// Check the issuer.
	if iss, ok := claims["iss"].(string); !ok || iss != "https://accounts.google.com" {
		return nil, fmt.Errorf("invalid ID token: %s", "iss")
	}

	clientId := ""
	webClientId := ""
	switch platformId {
	case constant.IOSPlatformID:
		clientId = config.Config.Oauth.Google.Ios.ClientID
		break
	case constant.AndroidPlatformID:
		clientId = config.Config.Oauth.Google.Android.ClientID
		webClientId = config.Config.Oauth.Google.Web.ClientID
		break
	}

	// Check the audience.
	if aud, ok := claims["aud"].(string); !ok || (aud != clientId && aud != webClientId) {
		return nil, fmt.Errorf("invalid ID token: %s", "aud")
	}

	// check expiry
	if exp, ok := claims["exp"].(float64); !ok || int64(exp) < time.Now().Unix() {
		return nil, fmt.Errorf("invalid ID token: %s", "exp")
	}

	return claims, nil
}
