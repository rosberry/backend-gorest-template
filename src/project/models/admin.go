package models

import (
	"time"
)

type (
	// Admin is the user model of admin panel.
	Admin struct {
		BaseModelWithSoftDelete
		Status uint
		Email  string
	}
)

const (
	AdminStatusNotConfirmed = iota
	AdminStatusConfirmed
)

func (admin *Admin) DisplayName() string {
	return admin.Email
}

func (admin *Admin) GetAccessExpireTime(link string) (expires time.Time) {
	return time.Now().Add(time.Hour * 24)
}
