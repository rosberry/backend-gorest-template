// nolint
package migrations

import (
	"time"

	"github.com/jinzhu/gorm"
)

// MIGRATION: create_admins_table
// use DBType to determine the type of DBMS

type M2020_05_25_080133 uint

var x2020_05_25_080133 = Add(M2020_05_25_080133(0))

func (m M2020_05_25_080133) String() string {
	return "create_admins_table"
}

func (m M2020_05_25_080133) DestructiveType() uint {
	return DestructiveDown
}

func (m M2020_05_25_080133) Up(tx *gorm.DB) error {
	type admin struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt *time.Time `sql:"index"`

		Status uint
		Email  string
	}

	err := tx.AutoMigrate(admin{}).Error
	return err
}

func (m M2020_05_25_080133) Down(tx *gorm.DB) error {
	return tx.DropTable("admins").Error
}
