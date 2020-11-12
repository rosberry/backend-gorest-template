// nolint
package migrations

import (
	"time"

	"github.com/jinzhu/gorm"
)

// MIGRATION: seed_default_admin
// use DBType to determine the type of DBMS

type M2020_11_11_154511 uint

var x2020_11_11_154511 = Add(M2020_11_11_154511(0))

func (m M2020_11_11_154511) String() string {
	return "seed_default_admin"
}

func (m M2020_11_11_154511) DestructiveType() uint {
	return DestructiveNo
}

func (m M2020_11_11_154511) Up(tx *gorm.DB) error {
	var approved uint = 1
	var email string = "admin@example.com"
	type admin struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt *time.Time `sql:"index"`

		Status uint
		Email  string
	}

	defaultAdmin := &admin{
		Email:  email,
		Status: approved,
	}

	err := tx.Save(defaultAdmin).Error
	if err != nil {
		return err
	}

	return tx.Exec("INSERT INTO auth_identities (provider, uid, encrypted_password, user_id) VALUES ('password', ?, '$2a$10$fov/LrnwMP0FL7lqpT2OLOBZc0a.oJbBJaagb5nacVcV/O0wmEUyu', ?);", email, defaultAdmin.ID).Error
}

func (m M2020_11_11_154511) Down(tx *gorm.DB) error {
	return Error("Method \"Down\" is not implemented!")
}
