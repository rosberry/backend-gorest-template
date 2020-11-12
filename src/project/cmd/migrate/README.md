# Gorm Migrator

Мигратор используется для создания и применения/отмены миграций. В своей работе использует [Gorm](https://gorm.io/).

## Добавление в проект

1. Мигратор копируется в качестве утилиты в проект. Как правило в поддиректорию `cmd/migrate`.
2. В файле `migrate.go` необходимо настроить пути импорта (заменить в двух местах `project` на название директории проекта).
3. Ожидается, что импортируемый пакет `models` экспортирует две функции:

```
// GetDB returns an instance of * gorm.DB
func GetDB() *gorm.DB

// GetDBType returns the type of the DBMS ("mysql" or "postgres") from config
func GetDBType() string
```

## Использование

### Первый запуск

Мигратор в своей работе опирается на модель проекта, которая для подключения к БД использует конфиг. Поэтому запуск следует проводить из такого места, чтобы файл конфига был доступен.

Некоторые возможные варианты запуска мигратора (предполагается, что `GOPATH` и прочее уже настроены):
* Файл `.env` расположен там же, где и `main.go` (package main). Тогда, находясь в этой директории, следует запустить команду `go run cmd/migrate/migrate.go -h` .
* Файл `.env` _уже_ расположен в директории `bin`. Можно собрать исполняемый файл с помощью команды `go install project/cmd/migrate` (`project` необходимо заменить на соответствующее название директории проекта). После этого из директории `bin` можно запускать: `./migrate -h` .


### Создание миграций

#### Изменения, связанные с переходом на gorm 2.0

**Во-первых**, все мигрирующие операции перенесены в отдельный интерфейс, доступ к которому можно получить с помощью функции `Migrator()`:

Вместо
```
tx.CreateTable(&cafes{})
```
необходимо использовать
```
tx.Migrator().CreateTable(&cafes{})
```

**Во-вторых**, в интерфейсе мигратора нет экспортируемой переменной `Error`, в которой хранится ошибка. Ошибка возвращается непосредственно при выполнении действия:

Вместо
```
return tx.CreateTable(&cafes{}).Error
```
необходимо использовать
```
return tx.Migrator().CreateTable(&cafes{})
```

**В-третьих**, была изменена сигнатура некоторых методов. Например, мигрирующих операций, связанных со столбцами:

Вместо
```
tx.Table("orders").DropColumn("status")
```
можно использовать
```
tx.Migrator().DropColumn(&Order{}, "status")
```

**В-четвертых**, удалена возможность явного добавления foreign keys. Теперь это нужно делать через определение модели:
Вместо
```
	type device struct {
		ID          uint `gorm:"primary_key"`
		UserID      uint   `gorm:"not null" sql:"DEFAULT:0"`
	}

	err := tx.AutoMigrate(device{}).Error
	if err == nil {
		err = tx.Model(&device{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE").Error
	}

	return err
```
нужно использовать
```
	type User struct {
		ID uint
	}

	type device struct {
		ID     uint `gorm:"primary_key"`
		UserID uint `gorm:"not null" sql:"DEFAULT:0"`
		User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}

	return tx.Migrator().AutoMigrate(device{})
```

### Просмотр списка

### Применение/отмена миграций

### Деплой

## Лицензия

© Rosberry, 2016

Опубликовано под лицензией [MIT](https://github.com/go-gorm/gorm/blob/master/License)
