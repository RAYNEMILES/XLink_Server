package utils

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io/ioutil"
	"log"
	"time"
)

var (
	PrivateKey *rsa.PrivateKey
)

func init() {
	privateBytes, err := ioutil.ReadFile("../rsa_private.pem")
	if err != nil {
		fmt.Println("Private Key: File Open Error.")
		return
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	if err != nil {
		panic(err)
	}
	PrivateKey = privateKey
}

type RequestDate struct {
	Phone       string `json:"phone"`
	InvitedCode string `json:"invite_code"`
	UserName    string `json:"username"`
	UserId      string `json:"username_id"`
}

type TokenClaims struct {
	Phone       string `json:"phone"`
	InvitedCode string `json:"invite_code"`
	UserName    string `json:"username"`
	UserId      string `json:"username_id"`
	jwt.StandardClaims
}

func GetDiscoverToken(userId string, InvitedCode string) string {
	data := &RequestDate{
		InvitedCode: InvitedCode,
		UserName:    "Im" + userId,
		UserId:      userId,
	}
	token, err := generateToken(data)
	if err != nil {
		return ""
	}
	return token
}

func generateToken(data *RequestDate) (string, error) {
	unixTime := time.Now().Unix()
	tokenExp := unixTime + 60*30 //30 min

	claims := TokenClaims{
		UserName:    data.UserName,
		InvitedCode: data.InvitedCode,
		UserId:      data.UserId,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  unixTime,
			ExpiresAt: tokenExp,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(PrivateKey)

	if err != nil {
		log.Println("Failed to sign id token string")
		return "", err
	}
	return ss, nil
}
