package db

import (
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"github.com/go-redis/redis/v9"
	"time"
)

const remoteTokenPrefix = "REMOTE_TOKEN:"

type RemoteTokenModel struct {
}

type RemoteTokenConsumeResult struct {
	UserID   string
	Platform int32
}

var RemoteTokenNotFoundError = errors.New("remote token not found")
var RemoteTokenNotNotAssignedError = errors.New("remote token not assigned")
var RemoteTokenInvalidSecretError = errors.New("invalid remote secret")

// Create generate new remote token
func (r RemoteTokenModel) Create() (string, string, error) {
	token := utils.RandomString(32)
	remoteSecret := utils.RandomString(32)

	key := remoteTokenPrefix + token

	if err := DB.rdb.HSet(context.Background(), key, "secret", remoteSecret).Err(); err != nil {
		return "", "", err
	}

	if err := DB.rdb.Expire(context.Background(), key, time.Minute).Err(); err != nil {
		return "", "", err
	}

	return token, remoteSecret, nil
}

// Assign add userID to token
func (r RemoteTokenModel) Assign(token string, userID string) error {
	key := remoteTokenPrefix + token

	if res, err := DB.rdb.Exists(context.Background(), key).Result(); err != nil {
		return err
	} else if res == 0 {
		return RemoteTokenNotFoundError
	}

	if res, err := DB.rdb.HSetNX(context.Background(), key, "userID", userID).Result(); err != nil {
		return err
	} else if res {
		return DB.rdb.Expire(context.Background(), key, time.Minute).Err()
	}

	return nil
}

// Consume drop key and return platform id
func (r RemoteTokenModel) Consume(token string, remoteSecret string) (string, error) {
	key := remoteTokenPrefix + token

	secret, err := DB.rdb.HGet(context.Background(), key, "secret").Result()
	if err == redis.Nil {
		return "", RemoteTokenNotFoundError
	} else if err != nil {
		return "", err
	}

	if secret != remoteSecret {
		return "", RemoteTokenInvalidSecretError
	}

	userID, err := DB.rdb.HGet(context.Background(), key, "userID").Result()
	if err == redis.Nil {
		return "", RemoteTokenNotNotAssignedError
	} else if err != nil {
		return "", err
	}

	if err = DB.rdb.Del(context.Background(), key).Err(); err != nil {
		return "", err
	}

	return userID, nil
}
