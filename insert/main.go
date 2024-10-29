package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

var authHeader = "Bearer 5Mxrg3TkCRq4aMy4PyO8QYA7BiWUqHy9fPlVbSruAlDpGj10ry4mgbbetL79M12S"
var baseUrl = "https://test.bible.clementineleaf.top/stress"

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
			insert(name, users)
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

func insert(name string, users []string) {
	tNow := time.Now()
	url := fmt.Sprintf("%s/%s/add", baseUrl, name)

	counter := 0
	for user, v := range timeIter(tNow.Add(-10*time.Hour), tNow, time.Millisecond, users) {
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
			req.Header.Add("Authorization", authHeader)
			client := &http.Client{}
			response, err := client.Do(req)

			if err != nil {
				slog.Error("bad request", "error", err.Error())
			} else if response.StatusCode != http.StatusNoContent {
				slog.Error("bad status code", "code", string(response.StatusCode))
			}

			if counter > 10_000 {
				slog.Info("insert")
			}
			counter++
		}
	}
}

func timeIter(tFrom, tTo time.Time, delta time.Duration, users []string) func(func(string, time.Time) bool) {

	return func(next func(uer string, timestamp time.Time) bool) {
		for iter := tFrom; iter.Before(tTo); iter = iter.Add(delta) {
			user := users[rand.IntN(len(users))]
			if !next(user, iter) {
				return
			}
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
