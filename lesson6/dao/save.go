package dao

import (
	"encoding/json"
	"os"
)

func SaveDB() error {
	data, err := json.Marshal(database)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/user_database.json", data, 0646)
	if err != nil {
		return err
	}
	return nil
}

func SaveRefreshToken() error {
	data, err := json.Marshal(refreshTokens)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/refresh_token.json", data, 0646)
	if err != nil {
		return err
	}
	return nil
}
