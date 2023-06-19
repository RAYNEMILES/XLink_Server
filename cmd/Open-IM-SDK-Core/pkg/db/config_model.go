package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
)

func (d *DataBase) GetLocalConfig(userID string) (*model_struct.LocalConfig, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var localConfig model_struct.LocalConfig
	err := d.conn.Where("user_id = ?", userID).Limit(1).Take(&localConfig).Error
	if err != nil {
		return nil, err
	}
	return &localConfig, nil
}

func (d *DataBase) SetLocalConfig(config *model_struct.LocalConfig) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var localConfig model_struct.LocalConfig
	err := d.conn.Where("user_id = ?", config.UserID).Limit(1).Take(&localConfig).Error
	if err != nil {
		err := d.conn.Create(&config).Error
		return err
	} else {
		err := d.conn.Where("user_id = ?", config.UserID).Updates(&config).Error
		return err
	}
}
