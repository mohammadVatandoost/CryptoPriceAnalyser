package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	// "strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/jinzhu/gorm"
)

type MetalPrice struct {
	gorm.Model
	Metal        string
	Price        float32
	TimeToSecond int64
	SignalAlg    int
	SignalAI     int
}

const metalAPIKey = "i9nq09yw30libb06we62w4x69j3pi75dl1b22vxvyse50s6f0847rom3inle"

// func main() {
// 	createMetalTable()
// 	updateMetalPrice()
// }

func createMetalTable() {
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("failed to connect database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&MetalPrice{})
}

func updateMetalPrice() {
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("updateMetalPrice failed to connect database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	resp, err := http.Get("https://metals-api.com/api/latest?access_key=i9nq09yw30libb06we62w4x69j3pi75dl1b22vxvyse50s6f0847rom3inle&base=USD&symbols=XAU")
	if err != nil {
		fmt.Println("updateMetalPrice failed to get metal data")
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("updateMetalPrice failed to ioutil.ReadAll(resp.Body)", err)
		os.Exit(1)
	}

	rates, _ := jsonparser.GetFloat(body, "rates", "XAU")
	goldPrice := 1.0 / rates
	fmt.Println("Gold Price per ounce:", goldPrice)
	db.Create(&MetalPrice{Metal: "XAU", Price: float32(goldPrice), TimeToSecond: time.Now().Unix(), SignalAlg: 0, SignalAI: 0})
}

func getGoldPrices() []MetalPrice {
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("failed to connect database: ", err)
		os.Exit(1)
	}
	defer db.Close()
	var goldPrice []MetalPrice
	db.Where("Metal = ? AND updated_at > ?", "XAU", time.Now().AddDate(0, 0, -7)).Find(&goldPrice)
	return goldPrice
}
