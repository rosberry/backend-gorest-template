// nolint
package migrations

import (
	"time"

	"github.com/jinzhu/gorm"
)

// MIGRATION: create_users_table
// use DBType to determine the type of DBMS

type M2020_05_14_074020 uint

var x2020_05_14_074020 = Add(M2020_05_14_074020(0))

func (m M2020_05_14_074020) String() string {
	return "create_users_table"
}

func (m M2020_05_14_074020) DestructiveType() uint {
	return DestructiveDown
}

func (m M2020_05_14_074020) Up(tx *gorm.DB) error {
	type user struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt *time.Time `sql:"index"`

		Name  string
		Photo string
	}
	return tx.CreateTable(&user{}).Error
}

func (m M2020_05_14_074020) Down(tx *gorm.DB) error {
	return tx.DropTable("users").Error
}
