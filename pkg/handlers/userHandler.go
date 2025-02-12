package handlers

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/goccy/go-json"

	"github.com/even44/JobsearchAPI/pkg/initializers"
	"github.com/even44/JobsearchAPI/pkg/models"
	"github.com/even44/JobsearchAPI/pkg/stores"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var UH *UserHandler

type UserHandler struct {
	store  stores.UserStore
	logger *log.Logger
}

func NewUserHandler(s stores.UserStore) *UserHandler {
	return &UserHandler{
		store:  s,
		logger: log.New(os.Stdout, "[USER] ", log.Ldate+log.Ltime+log.Lmsgprefix),
	}
}

func (h UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {

	// Get user credentials from body
	var body models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Printf("[ERROR] Received following error while parsing request JSON: \n%s", err)
		BadRequestHandler(w, r)
		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		InternalServerErrorHandler(w, r)
		h.logger.Printf("[ERROR] Could not hash password \n%s", err)
		return
	}

	// Create user
	user := models.User{Email: body.Email, Password: string(hash)}
	err = h.store.AddUser(&user)
	if err != nil {
		InternalServerErrorHandler(w, r)
		h.logger.Printf("[ERROR] Could not create user \n%s", err)
		return
	}

	// Respond

	w.WriteHeader(http.StatusOK)
}

func (h UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Get user credentials from body
	var body models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Printf("[ERROR] Received following error while parsing request JSON: \n%s", err)
		BadRequestHandler(w, r)
		return
	}

	//Look up user based on email
	user, err := h.store.GetUserByEmail(body.Email)
	if err != nil {
		InvalidEmailOrPasswordHandler(w, r)
		return
	}

	//user.Password is password hash in this case
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		InvalidEmailOrPasswordHandler(w, r)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(initializers.ApiSecret))

	if initializers.CookiesSecure {

		http.SetCookie(w, &http.Cookie{
			Name:        "Authorization",
			Value:       tokenString,
			MaxAge:      3600 * 24 * 30,
			HttpOnly:    true,
			Secure:      true,
			Partitioned: true,
			SameSite:    http.SameSiteNoneMode,
			Path:        "/auth",
		})
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:        "Authorization",
			Value:       tokenString,
			MaxAge:      3600 * 24 * 30,
			HttpOnly:    true,
			Secure:      false,
			Partitioned: false,
			SameSite:    http.SameSiteLaxMode,
			Path:        "/auth",
		})
	}
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	var loginResponse models.LoginResponse
	loginResponse.Email = user.Email

	jsonBytes, err := json.Marshal(loginResponse)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonBytes)
}
