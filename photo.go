package main

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// CreateDateTree
func CreateDateTree(dir string) (string, error) {

	tree := time.Now().Format("2006/01/02/")
	dirs := strings.Split(tree, "/")
	for x := 1; x < len(dirs); x++ {
		dirName := dir + strings.Join(dirs[:x], "/")
		if _, err := os.Stat(dirName); !os.IsNotExist(err) {
			continue
		}
		err := os.Mkdir(dirName, os.ModePerm)
		if err != nil {
			return dir, err
		}
	}

	return dir + tree, nil
}

// ++ work with photo

// Photo
type Photo struct {
	id        string
	name      string
	link      string
	dataTree  string
	directURL string
	dir       string
	Err       error
}

func (p *Photo) Id() string {
	return p.id
}

// Dir директория хранения
func (p *Photo) Dir() string {
	return p.dir
}

// Link ссылка на вебе
func (p *Photo) Link() string {
	return p.link + p.name
}

// Name имя файла
func (p *Photo) Name() string {
	return p.name
}

// Path полный путь
func (p *Photo) Path() string {
	return p.dir + p.name
}

// Exist проверка существования
func (p *Photo) Exist() bool {
	info, err := os.Stat(p.Path())
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Remove
func (p *Photo) Remove() error {
	err := os.Remove(p.Photof())
	if err != nil {
		return err
	}

	return nil
}

// Save сохранить исходный файл картинки
func (p *Photo) Save() *Photo {

	var photo = p.Photof()

	// проверяем существование, вдруг качали ранее
	p.name = p.id + ".jpeg"
	if p.Exist() {
		log.Printf("Файл %s найден!", photo)
		return p
	}

	// начинаем скачивание...
	resp, err := http.Get(p.directURL)
	if err != nil {
		return &Photo{Err: err}
	}

	// закрываем тело
	defer resp.Body.Close()

	// создаем файл
	f, err := os.Create(photo)
	if err != nil {
		return &Photo{Err: err}
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return &Photo{Err: err}
	}

	p.name = p.id + ".jpeg"

	return p
}

// photof() форм. имени картинки
func (p *Photo) Photof() string {
	return fmt.Sprintf("%s%s.jpeg", p.dir, p.id)
}

// NewDownloadPhoto
func NewDownloadPhoto(bot *tgbotapi.BotAPI, arr []tgbotapi.PhotoSize, dir, link string) (p *Photo) {

	if len(arr) > 0 {

		// создадим дерево каталогов по дате
		// и добавим к основному пути сохранения
		dir, _ := CreateDateTree(dir)
		link := link + time.Now().Format("2006/01/02/")
		// самая большая картинка
		fileID := arr[len(arr)-1].FileID

		// получаем ссылку на файл
		directURL, err := bot.GetFileDirectURL(fileID)
		if err != nil {
			return &Photo{
				Err: err,
			}
		}

		// проверяем существование каталога для картинок
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.Mkdir(dir, os.ModePerm)
		}

		return &Photo{
			id:        fileID,
			directURL: directURL,
			dir:       dir,
			link:      link,
		}
	}

	return nil
}

// UploadPhoto
func UploadPhoto(bot *tgbotapi.BotAPI, chatID int64, photo, text string) (tgbotapi.Message, error) {

	msg := tgbotapi.NewPhotoUpload(chatID, photo)
	msg.Caption = text
	msg.ParseMode = tgbotapi.ModeHTML

	message, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error upload photo message: chat id = %d , text = %s , error = %s", chatID, text, err)
	}

	return message, err
}

// -- work with photo
