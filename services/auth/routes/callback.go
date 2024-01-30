package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ferretcode-freelancing/fc-cloud/services/auth/session"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GithubResponse struct {
	AccessToken string `json:"access_token"`
}

type GithubUser struct {
	Id int `json:"id"`
}

type CloudUser struct {
	Id       int64         `json:"id"`
	Projects pq.Int64Array `json:"projects" gorm:"type:int[]"`
	Team     int64         `json:"team"`
}

func Callback(w http.ResponseWriter, r *http.Request, db gorm.DB, rdb redis.Client) error {
	code := r.URL.Query().Get("code")

	token, err := getCode(code)

	sid := uuid.NewString()

	cookie := http.Cookie{
		Name:   "fc-cloud",
		Value:  sid,
		Domain: os.Getenv("COOKIE_DOMAIN"),
		Path:   "/",
	}

	http.SetCookie(w, &cookie)

	if err != nil {
		return err
	}

	user, err := getUser(token)

	if err != nil {
		return err
	}

	err = session.CreateSession(sid, token, int64(user.Id), rdb)

	if err != nil {
		return err
	}

	err = createUser(user.Id, db)

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("You were successfully authenticated."))

	return nil
}

func createUser(githubId int, db gorm.DB) error {
	user := CloudUser{}

	err := db.First(&user, "id = ?", githubId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user := CloudUser{
				Id:       int64(githubId),
				Projects: pq.Int64Array([]int64{}),
				Team:     0,
			}

			err = db.Create(&user).Error

			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

func getUser(token string) (GithubUser, error) {
	client := http.Client{}

	req, err := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)

	if err != nil {
		return GithubUser{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	res, err := client.Do(req)

	if err != nil {
		return GithubUser{}, err
	}

	var user GithubUser

	parseErr := ProcessBody(res.Body, &user)

	if parseErr != nil {
		return GithubUser{}, parseErr
	}

	return user, nil
}

func getCode(code string) (string, error) {
	client := http.Client{}

	url := fmt.Sprintf(
		"https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_SECRET"),
		code,
	)

	req, err := http.NewRequest(
		"POST",
		url,
		nil,
	)

	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	var githubResponse GithubResponse

	parseErr := ProcessBody(res.Body, &githubResponse)

	if parseErr != nil {
		return "", parseErr
	}

	return githubResponse.AccessToken, nil
}

func ProcessBody(data io.ReadCloser, to interface{}) error {
	bytes, err := io.ReadAll(data)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &to); err != nil {
		return err
	}

	return nil
}
