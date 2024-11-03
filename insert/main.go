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

func NewLimiter(limitPerSecond int) Limiter {
	if limitPerSecond < 1 {
		panic("bro")
	}

	var limiter chan struct{}
	var auxLimit int
	var updateInterval time.Duration
	options := [7]time.Duration{
		time.Second,
		time.Millisecond * 100,
		time.Millisecond * 10,
		time.Millisecond,
		time.Microsecond * 100,
		time.Microsecond * 10,
		time.Microsecond,
	}
	for _, wait := range options {
		divider := int(time.Second / wait)
		prosPectLimit := limitPerSecond / divider

		if (limitPerSecond - prosPectLimit*divider) < 1 {
			updateInterval = wait
			auxLimit = prosPectLimit

			slog.Info("limiter conf try",
				"interval", wait,
				"max", auxLimit,
				"div", divider,
			)
		} else {
			limiter = make(chan struct{}, auxLimit)
			break
		}
	}

	go func() {
		for {
			gogo := true
			for cont := 0; gogo && cont < auxLimit; cont++ {
				select {
				case limiter <- struct{}{}:
				default:
					gogo = false
				}
			}
			time.Sleep(updateInterval)
		}
	}()
	return limiter
}

type config struct {
	Secret  string `env:"API_SECRET" envDefault:"shhhh"`
	BaseUrl string `env:"API_BASE_URL" envDefault:"http://localhost:9090"`
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
		NewLimiter(60),
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
	for v := range sendNPerTime(2, user, timeIter(refTime.Add(-10*time.Hour), refTime, 10*time.Millisecond)) {
		if data, err := json.Marshal(v); err != nil {
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
		url := fmt.Sprintf("%s/%s/list/add", baseUrl, name)
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

func sendNPerTime(n int, user string, iter func(func(time.Time) bool)) func(func([]Info) bool) {
	return func(next func([]Info) bool) {
		data := []Info{}

		for lTime := range iter {
			data = append(data, Info{
				User:      user,
				Timestamp: lTime,
				Data:      randomData(),
			})

			if len(data) == n {
				if !next(data) {
					return
				}
				data = []Info{}
			}
		}

		if len(data) > 0 {
			next(data)
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
	proxyLimiter := make(chan struct{}, 1)
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
