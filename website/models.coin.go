package main

import (
	// "strconv"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	// "net/http"
	"os"
	"time"

	// "github.com/buger/jsonparser"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const dbName = "Crypto.db"
const dbDriver = "sqlite3"

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

//   Inject fields `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` into model `User`
type Coin struct {
	gorm.Model
	Name     string
	CoinID   string
	Watching uint
}

type Price struct {
	gorm.Model
	CoinID    string
	Price     float32
	Volume    uint64
	MarketCap uint64
}

func getCoinPrice(coinID string) []Price {
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("failed to connect database: ", err)
		os.Exit(1)
	}
	defer db.Close()
	var prices []Price
	db.Where("coin_id = ? AND updated_at > ?", coinID, time.Now().AddDate(0, 0, -7)).Find(&prices)
	return prices
}

func getWatchingCoins() []Coin {
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("failed to connect database: ", err)
		os.Exit(1)
	}
	defer db.Close()
	var coins []Coin
	var watchingCoins []Coin
	db.Find(&coins)
	for i := 0; i < len(coins); i++ {
		if coins[i].Watching == 1 {
			watchingCoins = append(watchingCoins, coins[i])
		}
	}

	return watchingCoins
}
