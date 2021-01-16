package database

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pushbits/server/authentication/credentials"
	"github.com/pushbits/server/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database holds information for the database connection.
type Database struct {
	gormdb             *gorm.DB
	sqldb              *sql.DB
	credentialsManager *credentials.Manager
}

func createFileDir(file string) {
	if _, err := os.Stat(filepath.Dir(file)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
			panic(err)
		}
	}
}

// Create instanciates a database connection.
func Create(cm *credentials.Manager, dialect, connection string) (*Database, error) {
	log.Println("Setting up database connection.")

	maxOpenConns := 5

	var db *gorm.DB
	var err error

	switch dialect {
	case "sqlite3":
		createFileDir(connection)
		maxOpenConns = 1
		db, err = gorm.Open(sqlite.Open(connection), &gorm.Config{})
	case "mysql":
		db, err = gorm.Open(mysql.Open(connection), &gorm.Config{})
	default:
		message := "Database dialect is not supported"
		return nil, errors.New(message)
	}

	if err != nil {
		return nil, err
	}

	sql, err := db.DB()
	if err != nil {
		return nil, err
	}

	sql.SetMaxOpenConns(maxOpenConns)

	if dialect == "mysql" {
		sql.SetConnMaxLifetime(9 * time.Minute)
	}

	db.AutoMigrate(&model.User{}, &model.Application{})

	return &Database{gormdb: db, sqldb: sql, credentialsManager: cm}, nil
}

// Close closes the database connection.
func (d *Database) Close() {
	d.sqldb.Close()
}

// Populate fills the database with initial information like the admin user.
func (d *Database) Populate(name, password, matrixID string) error {
	var user model.User

	query := d.gormdb.Where("name = ?", name).First(&user)

	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		user := model.NewUser(d.credentialsManager, name, password, true, matrixID)

		if err := d.gormdb.Create(&user).Error; err != nil {
			return errors.New("user cannot be created")
		}
	} else {
		log.Printf("Admin user %s already exists.\n", name)
	}

	return nil
}
