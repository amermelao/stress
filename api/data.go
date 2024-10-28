package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type DBSecrets struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     int
}

func (s DBSecrets) String() string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s port=%d host=%s options='--client_encoding=UTF8'",
		s.User, s.Password, s.DBName, s.Port, s.Host,
	)
}

func NewConnection(s DBSecrets) (*gorm.DB, error) {
	return gorm.Open(
		postgres.New(postgres.Config{
			DSN:                  s.String(),
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{},
	)
}

type Info struct {
	User      string
	Data      map[string]string
	Timestamp time.Time
}

type Models interface {
	Info() Info
	New(Info) any
	fmt.Stringer
}

var tableNames = []any{&NoIndex{}, &TSV{}, &CreateAtUser{}}

func AddTables(db *gorm.DB) error {
	err := db.AutoMigrate(tableNames...)
	return err
}

func DropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(tableNames...)
}

func Insert[E Models](db *gorm.DB, data Info) error {
	var row E
	rowd := row.New(data)
	return db.Create(rowd).Error
}

func Get[S ~[]E, E Models](db *gorm.DB, _ string, from, to time.Time) ([]Info, error) {
	var data S
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("timestamp > ? and timestamp < ?", from, to).Find(&data)
	}

	fmt.Println(db.ToSQL(query))
	err := query(db).Error
	if err != nil {
		return nil, err
	}
	var returnInfo = []Info{}
	for _, a := range data {
		returnInfo = append(returnInfo, a.Info())
	}
	return returnInfo, err
}