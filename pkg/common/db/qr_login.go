package db

import (
	"context"
	"time"
)

func (d *DataBases) QrCodeSaveInfo(qrCodeId string, info map[string]string) bool {
	key := QrLoginInfo + qrCodeId
	defer d.rdb.Expire(context.Background(), key, time.Second*119)
	return d.rdb.HMSet(context.Background(), key, info).Val()
}

func (d *DataBases) QrCodeGetInfo(qrCodeId string) (map[string]string, error) {
	key := QrLoginInfo + qrCodeId
	return d.rdb.HGetAll(context.Background(), key).Result()
}
