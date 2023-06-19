package oauth

import (
	"Open_IM/pkg/common/config"
	"encoding/base64"
	"errors"
	"github.com/Timothylock/go-signin-with-apple/apple"
	"time"
)

func ParseAppleCode(platformId int, IDToken string, RedirectURI string) (map[string]interface{}, error) {
	decodeString, err := base64.StdEncoding.DecodeString(IDToken)
	if err != nil {
		return nil, err
	}
	appleUniqueId, err := apple.GetUniqueID(string(decodeString))
	if err != nil {
		return nil, err
	}

	claim, _ := apple.GetClaims(string(decodeString))
	email, _ := claim.Get("email")
	exp, _ := claim.Get("exp")
	aud, _ := claim.Get("aud")

	if exp.(float64) < float64(time.Now().Unix()) {
		return nil, errors.New("token expired")
	}
	if aud.(string) != config.Config.Oauth.Apple.AppleId {
		return nil, errors.New("token audience is not valid")
	}

	return map[string]interface{}{
		"id":    appleUniqueId,
		"email": email,
	}, nil
}
