package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v11"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

var rateLimiter <-chan bool

func limit(wait time.Duration, limit int) <-chan bool {
	limiter := make(chan bool, limit)

	go func() {
		for {
			gogo := true
			for cont := 0; gogo && cont < limit; cont++ {
				select {
				case limiter <- true:
				default:
					gogo = false
				}
			}
			time.Sleep(wait)
		}
	}()
	return limiter
}

type config struct {
	Secret  string `env:"API_SECRET" envDefault:"shhhh"`
	BaseUrl string `env:"API_BASE_URL" envDefault:"https://example.com"`
}

var cfg config

func init() {
	if err := env.Parse(&cfg); err != nil {
		slog.Error("bad init", "error", err.Error())
	} else {
		slog.Info("config  loaded")
	}

	rateLimiter = limit(100*time.Millisecond, 6)

}

var testNames = []string{"noindex", "tsv", "createatuser"}

func main() {

	users := []string{}
	for range rand.IntN(100) + 1 {
		users = append(users, randSeq(rand.IntN(10)+5))
	}

	slog.Info("users")
	for _, user := range users {
		slog.Info(user)
	}

	var wg sync.WaitGroup
	for _, name := range testNames {
		wg.Add(1)
		go func() {
			insert(name, cfg.BaseUrl, users)
			wg.Done()
		}()
	}
	wg.Wait()
}

type Info struct {
	User      string
	Data      map[string]string
	Timestamp time.Time
}

func insert(name, baseUrl string, users []string) {
	tNow := time.Now()
	url := fmt.Sprintf("%s/%s/add", baseUrl, name)

	var wg sync.WaitGroup

	for _, user := range users {
		wg.Add(1)
		go func() {
			insertPerUser(user, url, tNow)
			wg.Done()
		}()
	}

	wg.Wait()
}

func insertPerUser(user, url string, refTime time.Time) {
	counter := 0

	for v := range timeIter(refTime.Add(-10*time.Hour), refTime, 10*time.Millisecond) {
		<-rateLimiter
		info := Info{
			User:      user,
			Timestamp: v,
			Data:      randomData(),
		}

		if data, err := json.Marshal(info); err != nil {
			slog.Error("fail to encode data",
				"error", err.Error(),
			)
		} else if req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data)); err != nil {
			slog.Error("fail to send",
				"error", err.Error(),
			)
		} else {
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.Secret))
			client := &http.Client{}
			response, err := client.Do(req)

			if err != nil {
				slog.Error("bad request", "error", err.Error())
			} else if response.StatusCode != http.StatusNoContent {
				slog.Error("bad status code", "code", response.StatusCode)
			}

			if counter > 10_000 {
				slog.Info("insert", "time", v.Format(time.RFC3339))
			}
			counter++
		}
	}
}

func timeIter(tFrom, tTo time.Time, delta time.Duration) func(func(time.Time) bool) {

	return func(next func(timestamp time.Time) bool) {
		for iter := tFrom; iter.Before(tTo); {
			if !next(iter) {
				return
			}
			jitter := time.Duration(rand.IntN(10) + 1)
			iter = iter.Add(delta * jitter)
		}
	}
}

var letters = []rune(" abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}

func randomData() map[string]string {
	keys := rand.IntN(20) + 1

	data := map[string]string{}

	for range keys {
		keyLength := rand.IntN(10) + 1
		key := randSeq(keyLength)
		valueLength := rand.IntN(100) + 1
		value := randSeq(valueLength)
		data[key] = value
	}

	return data
}
