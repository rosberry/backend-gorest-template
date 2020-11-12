// nolint
package migrations

import (
	"time"

	"github.com/jinzhu/gorm"
)

// MIGRATION: create_devices_table
// use DBType to determine the type of DBMS

type M2020_05_14_102737 uint

var x2020_05_14_102737 = Add(M2020_05_14_102737(0))

func (m M2020_05_14_102737) String() string {
	return "create_devices_table"
}

func (m M2020_05_14_102737) DestructiveType() uint {
	return DestructiveDown
}

func (m M2020_05_14_102737) Up(tx *gorm.DB) error {
	type device struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt *time.Time `sql:"index"`

		UserID      uint
		DeviceToken string
		Token       string
	}
	err := tx.AutoMigrate(device{}).Error
	if err == nil {
		err = tx.Model(&device{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE").Error
	}
	return err
}

func (m M2020_05_14_102737) Down(tx *gorm.DB) error {
	return tx.DropTable("devices").Error
}
