package oauth

import (
	"Open_IM/pkg/common/config"
	"github.com/huandu/facebook"
)

func ParseFaceBookAccessToken(accessToken string) (map[string]interface{}, error) {
	globalApp := facebook.New(config.Config.Oauth.Facebook.ClientID, config.Config.Oauth.Facebook.ClientSecret)

	session := globalApp.Session(accessToken)
	result, err := session.Get("/me", nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}
