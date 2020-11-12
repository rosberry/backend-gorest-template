// nolint
package migrations

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/jinzhu/gorm"
)

const TEMPLATE = `// nolint
package migrations

import (
	"github.com/jinzhu/gorm"
)

// MIGRATION: %s
// use DBType to determine the type of DBMS

type M%s uint

var x%s = Add(M%s(0))

func (m M%s) String() string {
	return "%s"
}

func (m M%s) DestructiveType() uint {
	return DestructiveNo
}

func (m M%s) Up(tx *gorm.DB) error {
	return Error("Method \"Up\" is not implemented!")
}

func (m M%s) Down(tx *gorm.DB) error {
	return Error("Method \"Down\" is not implemented!")
}
`

const (
	DestructiveNo = iota
	DestructiveDown
	DestructiveUp
	DestructiveFully
)

func GetPath() (dir string, ok bool) {
	_, filename, _, ok := runtime.Caller(0)
	dir = path.Dir(filename)
	return
}

func GetTemplate(name, date string) string {
	return fmt.Sprintf(TEMPLATE, name, date, date, date, date, name, date, date, date)
}

func Error(text string) error {
	return errors.New(text)
}

var DBType string

type M interface {
	Up(tx *gorm.DB) error
	Down(tx *gorm.DB) error
	String() string
	DestructiveType() uint
}

var Ms = make(map[string]M)

func Add(a M) *M {
	Ms[strings.Split(reflect.TypeOf(a).String(), ".")[1][1:]] = a
	return &a
}
