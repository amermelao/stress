package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
	"time"
)

type GeneralTestCase interface {
	Register(mux *http.ServeMux)
}
type TestCase[E Models] struct {
	Name string
	db   *gorm.DB
}

func NewTestCase[E Models](name string, db *gorm.DB) TestCase[E] {
	return TestCase[E]{Name: name, db: db}
}

func (t TestCase[E]) urlPrefix() string {
	return fmt.Sprintf("/%s", t.Name)
}
func (t TestCase[E]) Register(base *http.ServeMux) {
	handler := http.NewServeMux()
	handler.HandleFunc("GET /query", t.getQuery)
	handler.HandleFunc("POST /add", t.addElement)

	slog.Info("add", "path", t.Name)

	base.Handle(t.urlPrefix()+"/", http.StripPrefix(t.urlPrefix(), handler))
}

func (t TestCase[E]) getQuery(writer http.ResponseWriter, request *http.Request) {
	user := request.URL.Query().Get("user")
	fromStr := request.URL.Query().Get("from")
	toStr := request.URL.Query().Get("to")
	slog.Info("hi")

	if from, err := time.Parse(time.RFC3339, fromStr); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		slog.Error("bad from",
			"path", t.Name,
			"value", fromStr,
			"error", err.Error(),
		)
	} else if to, err := time.Parse(time.RFC3339, toStr); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		slog.Error("bad to",
			"path", t.Name,
			"value", toStr,
			"error", err.Error(),
		)
	} else if val, err := Get[[]E, E](t.db, user, from, to); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("bad get",
			"path", t.Name,
			"from", fromStr,
			"to", toStr,
			"user", user,
			"error", err.Error(),
		)
	} else if err := json.NewEncoder(writer).Encode(val); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("enode err", "error", err.Error())
	} else {
		slog.Info("ok")
	}
}

func (t TestCase[E]) addElement(writer http.ResponseWriter, request *http.Request) {
	info := Info{}
	if err := json.NewDecoder(request.Body).Decode(&info); err != nil {
		slog.Error("bad post",
			"path", t.Name,
			"error", err.Error(),
		)
		writer.WriteHeader(http.StatusBadRequest)
	} else {
		defer request.Body.Close()
		if err := Insert[E](t.db, info); err != nil {
			slog.Error("fail to insert",
				"error", err.Error(),
				"test", t.Name,
			)
			writer.WriteHeader(http.StatusInternalServerError)
		} else {
			writer.WriteHeader(http.StatusNoContent)
		}
	}

}
