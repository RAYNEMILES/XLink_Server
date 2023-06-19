package db

import (
	"context"
	"time"
)

const CAPTCHA = "CAPTCHA:"

type RedisStore struct {
}

// Set set a capt
func (r RedisStore) Set(id string, value string) error {
	key := CAPTCHA + id
	return DB.rdb.Set(context.Background(), key, value, time.Minute*3).Err()

}

// Get get a capt
func (r RedisStore) Get(id string, clear bool) string {
	key := CAPTCHA + id
	val, err := DB.rdb.Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	if clear {
		err := DB.rdb.Del(context.Background(), key).Err()
		if err != nil {
			return ""
		}
	}
	return val
}

// Verify verify a capt
func (r RedisStore) Verify(id, answer string, clear bool) bool {
	v := RedisStore{}.Get(id, clear)
	return v == answer
}
