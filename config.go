package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	IsFirstRun bool
	Bot        *Bot
	DataBase   *DataBase
}

// Bot
type Bot struct {
	Token string
	Users []int64
	Debug bool
}

// DataBase
type DataBase struct {
	User,
	Password,
	Base string
	Debug bool
}

func (db *DataBase) StringConn() string {
	return fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True", db.User, db.Password, db.Base)
}

func newConfig(path string) (*Config, error) {

	if path == "" {
		path = "./app.yaml"
	}

	// проверяем на месте ли файл
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{}, errors.New("Config file 'app.yaml' not found! " +
			"Create file or copy app.yaml.example\n")
	}

	var config Config
	// читаем содержимое
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return &Config{}, errors.New("Error read config file\n")
	}
	// пробуем провести анмаршалинг
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return &Config{}, errors.New("Error unmarshal config file\n")
	}

	return &config, nil
}
