package main

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type JSONB map[string]string

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

// gorm.Model definition
type Model struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	Timestamp time.Time
	UserName  string
	Data      JSONB  `gorm:"type:jsonb;default:'{}';not null"`
	TSV       string `gorm:"->;type:tsvector GENERATED ALWAYS AS (jsonb_to_tsvector('english', data, '[\"string\"]')) STORED"`
}

type NoIndex struct {
	Model
}

func (t NoIndex) Info() Info {
	return Info{
		User:      t.UserName,
		Data:      t.Data,
		Timestamp: t.Timestamp,
	}
}

func (t NoIndex) New(info Info) any {
	me := NoIndex{}
	me.UserName = info.User
	me.Data = info.Data
	me.Timestamp = info.Timestamp

	return &me
}

func (t NoIndex) NewList(data []Info) any {
	var list []NoIndex
	for _, v := range data {
		list = append(list, NoIndex{
			Model: Model{
				UserName:  v.User,
				Data:      v.Data,
				Timestamp: v.Timestamp,
			},
		})
	}
	return &list
}

func (t NoIndex) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

type TSV struct {
	Model
	TSV string `gorm:"->;type:tsvector GENERATED ALWAYS AS (jsonb_to_tsvector('english', data, '[\"string\"]')) STORED;index"`
}

func (t TSV) Info() Info {
	return Info{
		User:      t.UserName,
		Data:      t.Data,
		Timestamp: t.Timestamp,
	}
}

func (t TSV) New(info Info) any {
	me := TSV{}
	me.UserName = info.User
	me.Data = info.Data
	me.Timestamp = info.Timestamp

	return &me
}

func (t TSV) NewList(data []Info) any {
	var list []TSV
	for _, v := range data {
		list = append(list, TSV{
			Model: Model{
				UserName:  v.User,
				Data:      v.Data,
				Timestamp: v.Timestamp,
			},
		})
	}
	return &list
}

func (t TSV) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

type CreateAtUser struct {
	Model
	Timestamp time.Time `gorm:"index:idx_,CreateAtUserpriority:2"`
	UserName  string    `gorm:"index:idx_,CreateAtUserpriority:1"`
}

func (t CreateAtUser) Info() Info {
	return Info{
		User:      t.UserName,
		Data:      t.Data,
		Timestamp: t.Timestamp,
	}
}

func (t CreateAtUser) New(info Info) any {
	me := CreateAtUser{}
	me.UserName = info.User
	me.Data = info.Data
	me.Timestamp = info.Timestamp

	return &me
}

func (t CreateAtUser) NewList(data []Info) any {
	var list []CreateAtUser
	for _, v := range data {
		list = append(list, CreateAtUser{
			Model: Model{
				UserName:  v.User,
				Data:      v.Data,
				Timestamp: v.Timestamp,
			},
		})
	}
	return &list
}

func (t CreateAtUser) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}
