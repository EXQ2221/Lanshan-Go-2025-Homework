package dao

import (
	"encoding/json"
	"os"
)

const dbPath = "data/user_database.json"
const rfPath = "data/refresh_token.json"

func LoadDB() error {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &database)
	if err != nil {
		return err
	}
	return nil
}

func LoadRefreshToken() error {
	data, err := os.ReadFile(rfPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &refreshTokens)
	if err != nil {
		return err
	}
	return nil
}
