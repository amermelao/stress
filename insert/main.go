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

type Limiter <-chan struct{}

func NewLimiter(wait time.Duration, limit int) Limiter {
	limiter := make(chan struct{}, limit)

	go func() {
		for {
			gogo := true
			for cont := 0; gogo && cont < limit; cont++ {
				select {
				case limiter <- struct{}{}:
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
}

func main() {

	users := []string{}
	for range rand.IntN(100) + 1 {
		users = append(users, randSeqNoSpace(rand.IntN(10)+5))
	}

	slog.Info("users")
	for _, user := range users {
		slog.Info(user)
	}

	var testNames = []string{"noindex", "tsv", "createatuser"}
	insert(cfg.BaseUrl, testNames, users)
}

type Info struct {
	User      string
	Data      map[string]string
	Timestamp time.Time
}

func insert(baseUrl string, testCases, users []string) {
	tNow := time.Now()

	var wg sync.WaitGroup

	rateLimiter := LogPerXMessagesSend(
		NewLimiter(100*time.Millisecond, 20),
		2000,
	)

	for _, user := range users {
		wg.Add(1)
		go func() {
			insertPerUser(user, baseUrl, testCases, tNow, rateLimiter)
			wg.Done()
		}()
	}

	wg.Wait()
}

func insertPerUser(
	user, baseUrl string,
	testCases []string,
	refTime time.Time,
	rateLimiter Limiter,
) {
	for v := range timeIter(refTime.Add(-10*time.Hour), refTime, 10*time.Millisecond) {
		info := Info{
			User:      user,
			Timestamp: v,
			Data:      randomData(),
		}

		if data, err := json.Marshal(info); err != nil {
			slog.Error("fail to encode data",
				"error", err.Error(),
			)
		} else {
			submitJson(data, baseUrl, testCases, rateLimiter)
		}
	}
}

func submitJson(data []byte, baseUrl string, testCases []string, rateLimiter Limiter) {
	for _, name := range testCases {
		<-rateLimiter
		url := fmt.Sprintf("%s/%s/add", baseUrl, name)
		if req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data)); err != nil {
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

var lettersNoSpace = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeqNoSpace(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = lettersNoSpace[rand.IntN(len(lettersNoSpace))]
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

func LogPerXMessagesSend(limiter Limiter, perXMsg int) Limiter {
	proxyLimiter := make(chan struct{}, 5)
	go func() {
		count := 0
		for {
			v := <-limiter
			proxyLimiter <- v
			count++

			if count >= perXMsg {
				slog.Info("send messages",
					"value", perXMsg,
				)
				count = 0
			}
		}
	}()

	return proxyLimiter
}
