package mobile

import (
	"familycoin/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

var notAuth = []string{"/api/user/new", "/api/user/login"}

func JwtAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestPath := r.URL.Path

		for _, value := range notAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		response := make(map[string]interface{})
		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" {
			response = Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			Respond(w, response)
			return
		}

		arr := strings.Split(tokenHeader, " ")
		if len(arr) != 2 {
			response = Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			Respond(w, response)
			return
		}

		tk := &models.Token{}
		token, err := jwt.ParseWithClaims(arr[1], tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenPwd), nil
		})

		if err != nil {
			response = Message(false, "Malformed authentication token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			Respond(w, response)
			return
		}

		if !token.Valid {
			response = Message(false, "Token is not valid.")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			Respond(w, response)
			return
		}

		ctx := context.WithValue(r.Context(), "telegram_id", tk.TelegramId)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func checkLogin(login, password string) map[string]interface{} {

	account := models.User{}
	account.Login = login
	err := account.Read()

	if err != nil || account.ID == 0 {
		return Message(false, "Invalid login credentials. Please try again")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))

	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return Message(false, "Invalid login credentials. Please try again")
	}

	account.Password = ""

	// Create JWT token
	tk := &models.Token{UserId: account.ID, TelegramId: account.TelegramId}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(tokenPwd))
	account.Token = tokenString

	resp := Message(true, "Logged In")
	resp["account"] = account

	return resp
}
