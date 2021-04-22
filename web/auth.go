package web

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"familycoin/models"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var BotToken string
var BotName string
var SessionLife time.Duration

type Sessions struct {
	sync.Mutex
	Map    map[string]*models.User
	Ticker map[string]*time.Timer
}

var sessions = &Sessions{
	Map:    map[string]*models.User{},
	Ticker: map[string]*time.Timer{},
}

type notAuth struct {
	Mask *regexp.Regexp
}

var notAuthPath = []notAuth{
	{Mask: regexp.MustCompile(".*t.me.*")},
	{Mask: regexp.MustCompile("/singin")},
	{Mask: regexp.MustCompile("/login.*")},
	{Mask: regexp.MustCompile("/static/.*")},
}

func auth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		requestPath := request.URL.Path

		for i := range notAuthPath {

			search := notAuthPath[i].Mask.FindStringSubmatch(requestPath)

			if len(search) != 0 {
				next.ServeHTTP(writer, request)
				return
			}
		}

		token, err := request.Cookie("_token")
		if err != nil {
			logger.Printf("Error Read cookie session token %v", err)
			http.Redirect(writer, request, "/singin", 302)
			return
		}

		if _, ok := sessions.Map[token.Value]; !ok {
			logger.Printf("user not login")
			http.Redirect(writer, request, "/singin", 302)
			return
		}

		ctx := context.WithValue(request.Context(), "token", token.Value)
		request = request.WithContext(ctx)

		next.ServeHTTP(writer, request)
	})
}

func checkTelegramLogin(request *http.Request) (bool, string) {

	// return true, "123"
	if _, ok := request.URL.Query()["hash"]; !ok {
		return false, ""
	}

	keyHash := sha256.New()
	keyHash.Write([]byte(BotToken))
	secretkey := keyHash.Sum(nil)

	var checkparams []string
	for k, v := range request.URL.Query() {
		if k != "hash" {
			checkparams = append(checkparams, fmt.Sprintf("%s=%s", k, v[0]))
		}
	}
	sort.Strings(checkparams)

	checkString := strings.Join(checkparams, "\n")
	hash := hmac.New(sha256.New, secretkey)
	hash.Write([]byte(checkString))
	hashstr := hex.EncodeToString(hash.Sum(nil))

	return hashstr == request.URL.Query()["hash"][0], request.URL.Query()["hash"][0]
}

func login(writer http.ResponseWriter, request *http.Request) {

	id := request.URL.Query()["id"][0]
	telegramId, _ := strconv.Atoi(id)

	if u := models.GetUser(int64(telegramId)); u.ID != 0 {

		ok, hash := checkTelegramLogin(request)
		if !ok {
			http.Redirect(writer, request, "/singin", 302)
			return
		}

		sessions.Lock()

		sessions.Map[hash] = u
		sessions.Ticker[hash] = time.AfterFunc(SessionLife, func() {
			delete(sessions.Map, hash)
			delete(sessions.Ticker, hash)
		})
		SetCookie(writer, "_token", hash, SessionLife)

		sessions.Unlock()

		http.Redirect(writer, request, "/", 302)

		return
	}

	http.Redirect(writer, request, "/singin", 302)

	return

}

func SetCookie(w http.ResponseWriter, name, value string, ttl time.Duration) error {

	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:       name,
		Value:      value,
		Domain:     "/",
		Path:       "/",
		Expires:    expire,
		RawExpires: expire.Format(time.RFC3339),
	}
	http.SetCookie(w, &cookie)

	return nil
}
