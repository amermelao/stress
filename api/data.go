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
	db, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN:                  s.String(),
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{},
	)

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(200)

	return db, err
}

type Info struct {
	User      string
	Data      map[string]string
	Timestamp time.Time
}

type Models interface {
	Info() Info
	New(Info) any
	NewList([]Info) any
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

func InsertMany[E Models](db *gorm.DB, data []Info) error {
	var row E
	query := func(db *gorm.DB) *gorm.DB {
		return db.Create(row.NewList(data))
	}
	return query(db).Error
}

func Get[S ~[]E, E Models](db *gorm.DB, user string, from, to time.Time) ([]Info, error) {
	var data S
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("user_name = ? and timestamp > ? and timestamp < ?", user, from, to).Find(&data)
	}
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

func Users[E Models](db *gorm.DB) ([]string, error) {
	results := []string{}
	err := db.Model(new(E)).Distinct("user_name").Find(&results).Error
	return results, err
}
