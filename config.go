package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Config
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

// stringConn format string connection database
func (db *DataBase) stringConn() string {
	return fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True", db.User, db.Password, db.Base)
}

// newConfig reading an unmarshal app.yaml
func newConfig() (*Config, error) {

	path := "./app.yaml"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{}, errors.New("Config file 'app.yaml' not found! " +
			"Create file or copy app.yaml.example\n")
	}

	var config Config
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return &Config{}, errors.New("Error read config file\n")
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return &Config{}, errors.New("Error unmarshal config file\n")
	}

	return &config, nil
}
