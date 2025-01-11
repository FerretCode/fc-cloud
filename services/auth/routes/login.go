package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Login(w http.ResponseWriter, r *http.Request, db gorm.DB, rdb redis.Client) error {
	url := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&scope=read:user&redirect_uri=%s://%s:%s/auth/callback",
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("CALLBACK_URL_PROTOCOL"),
		os.Getenv("CALLBACK_URL_HOST"),
		os.Getenv("CALLBACK_URL_PORT"),
	)

	http.Redirect(w, r, url, http.StatusFound)

	return nil
}
