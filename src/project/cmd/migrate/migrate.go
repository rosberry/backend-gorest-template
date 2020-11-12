// nolint
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	m "project/cmd/migrate/migrations"
	"project/models"
)

const APP_VERSION = "0.2"

type Migration struct {
	Migration string `sql:"unique_index" json:"migration"`
	Batch     uint   `json:"batch"`
}

type Deployment struct {
	DeploymentID string `sql:"unique_index" json:"deployment_id"`
	InstanceID   string `json:"instance_id"`
	Status       uint   `json:"status"`
}

const (
	DeploymentStatusInProgress = iota
	DeploymentStatusSuccess
	DeploymentStatusFail
)

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var newMigration *string = flag.String("new", "", "Create a new migration.")
var upFlag *bool = flag.Bool("up", false, "Execute all of your migrations.")
var downFlag *bool = flag.Bool("down", false, "Rollback the latest migration operation.")
var oneFlag *bool = flag.Bool("one", false, "Only one migration.")
var listFlag *bool = flag.Bool("list", false, "Show a list of migrations.")
var forceFlag *bool = flag.Bool("force", false, "Forcing migrations.")
var sFlag *bool = flag.Bool("s", false, "Single transaction.")
var deployFlag *bool = flag.Bool("deploy", false, "Use to deploy.")

var deployment, instance string

const (
	fmtOk      = "32"
	fmtNew     = "36"
	fmtNotice  = "1;33"
	fmtWarning = "1;31"
	fmtError   = "7;91"
	fmtInfo    = "1;37"
	fmtNotInv  = "7;93"
)

func fmtPrint(format, template string, params ...interface{}) {
	fmt.Printf("\033[%sm", format)
	fmt.Printf(template, params...)
	fmt.Printf("\033[0;39m")
}

const (
	migrStatusOk = iota
	migrStatusNew
	migrStatusNotFound
	migrStatusMissing
)

type migration struct {
	Date   string
	Name   string
	Batch  uint
	Status uint
}

const (
	dateFormat    = "2006_01_02_150405"
	dateFormatLen = len(dateFormat)
)

func createMigration(name string) {
	if dir, ok := m.GetPath(); ok {
		name = strings.TrimSpace(name)
		if strings.ToLower(name[len(name)-4:]) == "test" {
			fmtPrint(fmtWarning, "Migration's name cannot end with a \"test\"!\n")
			os.Exit(1)
		}
		migrTime := time.Now().Format(dateFormat)
		migrName := strings.Replace(name, " ", "_", -1)
		fileName := migrTime + "_" + migrName
		file, err := os.Create(dir + "/" + fileName + ".go")
		if err != nil {
			fmtPrint(fmtWarning, "File can't create...")
			os.Exit(1)
		}
		defer file.Close()

		file.WriteString(m.GetTemplate(migrName, migrTime))
		fmtPrint(fmtInfo, "Migration %s was created\n", fileName)
		os.Exit(0)
	} else {
		fmtPrint(fmtWarning, "Path to migrations can't be obtained...")
		os.Exit(1)
	}
}

func getInstanceId() (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return string(body), nil
}

func setDeployStatus(tx *gorm.DB, status uint) {
	tx.Model(Deployment{}).Where("deployment_id = ?", deployment).
		Updates(Deployment{Status: status})
	tx.Commit()
}

func main() {
	flag.Parse() // Scan the arguments list

	switch {
	case *versionFlag:
		fmt.Println("Version:", APP_VERSION)
		os.Exit(0)
	case *newMigration != "":
		createMigration(*newMigration)
	case *upFlag, *downFlag, *listFlag:
		if *oneFlag && *sFlag {
			fmtPrint(fmtWarning, "Flags \"one\" and \"s\" can not be used together...\n")
			os.Exit(2)
		}
		if *deployFlag {
			deployment = os.Getenv("DEPLOYMENT_ID")
			if deployment == "" {
				fmtPrint(fmtWarning, "Environment variable \"DEPLOYMENT_ID\" not found...\n")
				os.Exit(1)
			}
			var err error
			instance, err = getInstanceId()
			if err != nil {
				fmtPrint(fmtWarning, "Failed to get an instance_id...\n")
				os.Exit(1)
			}
		}
		db := models.GetDB()
		defer db.Close()
		tableOptions := ""
		m.DBType = models.GetDBType()
		if m.DBType == "mysql" {
			db.Exec("SET autocommit = 0;")
			tableOptions = "DEFAULT CHARSET=utf8"
		}
		db.Set("gorm:table_options", tableOptions).
			AutoMigrate(&Migration{}, &Deployment{})

		var tdep *gorm.DB
		if *deployFlag {
			tdep = db.Begin()
			err := tdep.Create(Deployment{deployment, instance, DeploymentStatusInProgress}).Error
			if err != nil {
				tdep.Rollback()
				var dep Deployment
				db.Where("deployment_id = ?", deployment).First(&dep)
				if dep.Status == DeploymentStatusSuccess {
					fmtPrint(fmtInfo, "Batch migrations has been successfully completed on an instance %s\n", instance)
					os.Exit(0)
				} else {
					fmtPrint(fmtError, "Batch migrations has been failed on an instance %s\n", instance)
					os.Exit(1)
				}
			}
		}

		var dbMigrs []Migration
		db.Order("batch, migration").Find(&dbMigrs)

		migrations := make([]migration, 0, len(dbMigrs)+len(m.Ms))
		var status uint
		for _, v := range dbMigrs {
			status = migrStatusNotFound
			if _, ok := m.Ms[v.Migration[:dateFormatLen]]; ok {
				status = migrStatusOk
			}
			migrations = append(migrations, migration{v.Migration[:dateFormatLen], v.Migration[dateFormatLen+1:], v.Batch, status})
		}
		last := len(migrations) - 1

		newMigrs := make([]string, 0, len(m.Ms))
	OuterLoop:
		for k := range m.Ms {
			for _, v := range migrations {
				if v.Date == k {
					continue OuterLoop
				}
			}
			newMigrs = append(newMigrs, k)
		}
		sort.Strings(newMigrs)
		for _, v := range newMigrs {
			if last < 0 || v > migrations[last].Date {
				migrations = append(migrations, migration{v, m.Ms[v].String(), 0, migrStatusNew})
			} else {
				migrations = append(migrations, migration{v, m.Ms[v].String(), 0, migrStatusMissing})
			}
		}

		switch {
		case *listFlag:
			if len(migrations) == 0 {
				fmtPrint(fmtInfo, "No migrations\n")
			} else {
				fmt.Println()
				var fmtMsg, fmtSign, sign string
				var batch uint = migrations[0].Batch
				for i, v := range migrations {
					if i == 0 && last >= 0 {
						fmtPrint(fmtInfo, "Implemented migrations:\n")
					}
					if batch != v.Batch {
						fmt.Println()
						batch = v.Batch
					}
					if i == last+1 {
						fmtPrint(fmtInfo, "New migrations:\n")
					}
					switch v.Status {
					case migrStatusOk:
						fmtMsg = fmtOk
						fmtSign = fmtOk
						sign = " "
					case migrStatusNew:
						fmtMsg = fmtNew
						fmtSign = fmtNew
						sign = " "
					case migrStatusNotFound:
						fmtMsg = fmtWarning
						fmtSign = fmtError
						sign = "!"
					case migrStatusMissing:
						fmtMsg = fmtNotice
						fmtSign = fmtNotInv
						sign = "!"
					}
					fmtPrint(fmtMsg, "%s ", v.Date)
					fmtPrint(fmtSign, sign)
					fmtPrint(fmtMsg, "  %s\n", v.Name)
				}
				fmt.Println()
			}
		case *upFlag:
			if last+1 == len(migrations) {
				fmtPrint(fmtInfo, "No new migrations...\n")
			} else {
				var newBatch uint = 1
				if last >= 0 {
					newBatch = migrations[last].Batch + 1
				}
				var count uint
				var tx *gorm.DB
				if *sFlag {
					tx = db.Begin()
				}
				for i := last + 1; i < len(migrations); i++ {
					migrName := migrations[i].Date + "_" + migrations[i].Name
					switch m.Ms[migrations[i].Date].DestructiveType() {
					case m.DestructiveUp, m.DestructiveFully:
						if !*forceFlag {
							if *sFlag {
								tx.Rollback()
							}
							fmtPrint(fmtError, "Destructive UP of migration %s!\n", migrName)
							if *sFlag {
								fmtPrint(fmtWarning, "All migrations in this batch have been canceled!\n")
							}
							fmtPrint(fmtInfo, "Use -force for executing.\n")
							if *deployFlag {
								setDeployStatus(tdep, DeploymentStatusFail)
							}
							os.Exit(1)
						}
					}
					if !*sFlag {
						tx = db.Begin()
					}
					err := m.Ms[migrations[i].Date].Up(tx)
					if err != nil {
						tx.Rollback()
						fmtPrint(fmtError, "Error during execution UP of migration %s:\n", migrName)
						fmtPrint(fmtWarning, "%s\n", err.Error())
						if *sFlag {
							fmtPrint(fmtWarning, "All migrations in this batch have been canceled!\n")
						}
						if *deployFlag {
							setDeployStatus(tdep, DeploymentStatusFail)
						}
						os.Exit(1)
					}
					tx.Create(Migration{migrName, newBatch})
					if !*sFlag {
						tx.Commit()
						db.Model(&Migration{}).Where("migration = ?", migrName).Count(&count)
						if count == 0 {
							fmtPrint(fmtError, "TRANSACTION was ABORTED during execution UP of migration %s !!!\n", migrName)
							if *deployFlag {
								setDeployStatus(tdep, DeploymentStatusFail)
							}
							os.Exit(1)
						}
					}
					fmtPrint(fmtOk, "UP %s -- Ok\n", migrName)
					if *oneFlag {
						break
					}
				}
				if *sFlag {
					tx.Commit()
					db.Model(&Migration{}).Where("batch = ?", newBatch).Count(&count)
					if count == 0 {
						fmtPrint(fmtError, "TRANSACTION was ABORTED during execution UP of migrations !!!\n")
						fmtPrint(fmtWarning, "All migrations in this batch have been canceled!\n")
						if *deployFlag {
							setDeployStatus(tdep, DeploymentStatusFail)
						}
						os.Exit(1)
					}
				}
			}
		case *downFlag:
			if last < 0 {
				fmtPrint(fmtInfo, "No implemented migrations...\n")
			} else {
				var count uint
				var tx *gorm.DB
				if *sFlag {
					tx = db.Begin()
				}
				for i := last; i >= 0; i-- {
					if migrations[i].Batch != migrations[last].Batch {
						break
					}
					migrName := migrations[i].Date + "_" + migrations[i].Name
					if migrations[i].Status == migrStatusNotFound {
						if *sFlag {
							tx.Rollback()
						}
						fmtPrint(fmtError, "NOT FOUND migration %s !!!\n", migrName)
						if *sFlag {
							fmtPrint(fmtWarning, "All migrations in this batch have been reverted!\n")
						}
						if *deployFlag {
							setDeployStatus(tdep, DeploymentStatusFail)
						}
						os.Exit(1)
					}
					switch m.Ms[migrations[i].Date].DestructiveType() {
					case m.DestructiveDown, m.DestructiveFully:
						if !*forceFlag {
							if *sFlag {
								tx.Rollback()
							}
							fmtPrint(fmtError, "Destructive DOWN of migration %s!\n", migrName)
							if *sFlag {
								fmtPrint(fmtWarning, "All migrations in this batch have been reverted!\n")
							}
							fmtPrint(fmtInfo, "Use -force for executing.\n")
							if *deployFlag {
								setDeployStatus(tdep, DeploymentStatusFail)
							}
							os.Exit(1)
						}
					}
					if !*sFlag {
						tx = db.Begin()
					}
					err := m.Ms[migrations[i].Date].Down(tx)
					if err != nil {
						tx.Rollback()
						fmtPrint(fmtError, "Error during execution DOWN of migration %s:\n", migrName)
						fmtPrint(fmtWarning, "%s\n", err.Error())
						if *sFlag {
							fmtPrint(fmtWarning, "All migrations in this batch have been reverted!\n")
						}
						if *deployFlag {
							setDeployStatus(tdep, DeploymentStatusFail)
						}
						os.Exit(1)
					}
					tx.Delete(Migration{}, "migration = ?", migrName)
					if !*sFlag {
						tx.Commit()
						db.Model(&Migration{}).Where("migration = ?", migrName).Count(&count)
						if count > 0 {
							fmtPrint(fmtError, "TRANSACTION was ABORTED during execution DOWN of migration %s !!!\n", migrName)
							if *deployFlag {
								setDeployStatus(tdep, DeploymentStatusFail)
							}
							os.Exit(1)
						}
					}
					fmtPrint(fmtOk, "DOWN %s -- Ok\n", migrName)
					if *oneFlag {
						break
					}
				}
				if *sFlag {
					tx.Commit()
					db.Model(&Migration{}).Where("batch = ?", migrations[last].Batch).Count(&count)
					if count > 0 {
						fmtPrint(fmtError, "TRANSACTION was ABORTED during execution DOWN of migrations !!!\n")
						fmtPrint(fmtWarning, "All migrations in this batch have been reverted!\n")
						if *deployFlag {
							setDeployStatus(tdep, DeploymentStatusFail)
						}
						os.Exit(1)
					}
				}
			}
		}
		if *deployFlag {
			setDeployStatus(tdep, DeploymentStatusSuccess)
		}
		os.Exit(0)
	default:
		fmt.Println("Use -h for help.")
	}
}
