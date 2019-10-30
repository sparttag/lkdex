package dex

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func connectDB(driverName string, dbName string) (*SQLDBBackend, error) {
	db, err := gorm.Open(driverName, dbName)
	if err != nil {
		return nil, err
	}
	return &SQLDBBackend{*db}, nil
}

type User struct {
	gorm.Model
	Age uint
}

func TestDB(t *testing.T) {
	db, err := connectDB("sqlite3", "order.db")
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&User{})
	//dex.SaveOrder()
	err = db.Create(&User{Age: 123}).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.Create(&User{Age: 1111}).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.Create(&User{Age: 456}).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.Create(&User{Age: 123}).Error
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.Create(&User{Model: gorm.Model{ID: 123}}).Error
	if err != nil {
		fmt.Println(err.Error())
	}

	err = db.Create(&User{Model: gorm.Model{ID: 123}}).Error
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestDBOrder(t *testing.T) {
	db, err := connectDB("sqlite3", "order.db")
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&OrderModel{})
	//db.CreateOrder()
}

func TestAccount(t *testing.T) {
	db, err := connectDB("sqlite3", "order.db")
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate()
}

func TestBlockSync(t *testing.T) {
	db, err := connectDB("sqlite3", "order.db")
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate()
}

func TestTrade(t *testing.T) {
	db, err := connectDB("sqlite3", "order.db")
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate()
}
