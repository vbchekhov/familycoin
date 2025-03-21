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
	Bot      *Bot      `yaml:"bot"`
	DataBase *DataBase `yaml:"database"`
	Web      *Web      `yaml:"web"`
	Mobile   *Mobile   `yaml:"mobile"`
}

// Bot
type Bot struct {
	Token string  `yaml:"token"`
	Name  string  `yaml:"name"`
	Users []int64 `yaml:"users"`
	Owner int64   `yaml:"owner"`
	Debug bool    `yaml:"debug"`
}

// DataBase
type DataBase struct {
	Host     string `yaml:"host"`
	Post     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Base     string `yaml:"base"`
	Debug    bool   `yaml:"debug"`
}

// ConnToMariaDB format string connection database
func (db *DataBase) ConnToMariaDB() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", db.User, db.Password, db.Host, db.Post, db.Base)
}

type Web struct {
	Port    string `yaml:"port"`
	CertSRT string `yaml:"certSRT"`
	CertKEY string `yaml:"certKEY"`
	Debug   bool   `yaml:"debug"`
}

func (w *Web) Portf() string {
	return ":" + w.Port
}

func (w *Web) IsTSL() bool {
	return w.CertKEY != "" && w.CertSRT != ""
}

type Mobile struct {
	Port     string `yaml:"port"`
	TokenPwd string `yaml:"token_pwd"`
	CertSRT  string `yaml:"certSRT"`
	CertKEY  string `yaml:"certKEY"`
	Debug    bool   `yaml:"debug"`
}

func (m *Mobile) Portf() string {
	return ":" + m.Port
}

func (m *Mobile) IsTSL() bool {
	return m.CertKEY != "" && m.CertSRT != ""
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
		return &Config{}, errors.New("Error Read config file\n")
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return &Config{}, errors.New("Error unmarshal config file\n")
	}

	return &config, nil
}
