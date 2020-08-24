package main

import (
	"strconv"
	// "encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/buger/jsonparser"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// const dbName = "Crypto.db"
// const dbDriver = "sqlite3"
const APIkey = "3a6d281a390561754bd63457b8e5d904"
const numberOfRoutines = 4
const serverPort = "4567"
const NoConnection = "NoConnection"

// // gorm.Model definition
// type Model struct {
// 	ID        uint `gorm:"primary_key"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	DeletedAt *time.Time
// }

// //   Inject fields `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` into model `User`
// type Coin struct {
// 	gorm.Model
// 	Name     string
// 	CoinID   string
// 	Watching uint
// }

// type Price struct {
// 	gorm.Model
// 	CoinID    string
// 	Price     float32
// 	Volume    uint64
// 	MarketCap uint64
// }

type coinJSONObject struct {
	id   string
	name string
}

func main() {
	// createCoinTable()
	// enterCoinsInfo()
	// enterCoinPrice()
	// analysePrice()

	// ticker := time.NewTicker(60 * time.Second)
	// done := make(chan bool)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-done:
	// 			ticker.Stop()
	// 			return
	// 		case <-ticker.C:
	// enterCoinPrice()
	// analysePrice()
	// fmt.Println("=================================")
	// 		}
	// 	}
	// }()

	// done <- true
	// fmt.Println("Server Port:", serverPort)
	// http.HandleFunc("/", requestHandler)
	// http.ListenAndServe(":"+serverPort, nil)

	go initalizeRoutes()
	enterCoinPrice()
	analysePrice()
	fmt.Println("=================================")

	for range time.Tick(time.Minute * 5) {
		enterCoinPrice()
		analysePrice()
		fmt.Println("=================================")
	}

	for range time.Tick(time.Hour * 8) {
		updateMetalPrice()
	}

	fmt.Println("Ticker stopped")
}

func enterCoinsInfo() {
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("enterCoinsInfo failed to connect database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	resp, err := http.Get("https://api.nomics.com/v1/currencies?key=" + APIkey)
	if err != nil {
		fmt.Println("failed to get coin data")
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("enterCoinsInfo failed to ioutil.ReadAll(resp.Body)", err)
		os.Exit(1)
	}

	fmt.Println(string(body))

	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		name, _ := jsonparser.GetString(value, "name")
		id, _ := jsonparser.GetString(value, "id")
		fmt.Println(" name:", name, ", id:", id, "")
		db.Create(&Coin{Name: name, CoinID: id, Watching: 0})
	})

	// var coins []coinJSONObject
	// var coins map[string]interface{}
	// var coins []string
	// var coins []map[string]interface{}
	// json.Unmarshal(body, &coins)
	// for i := range coins {
	// 	for key, value := range coins[i] {
	// 		// Each value is an interface{} type, that is type asserted as a string
	// 		fmt.Println(key, value.(string))
	// 	}
	// 	// println("Coin ", i, " :", coins[i], ", ", coins[i]["name"])
	// 	// println("Coin ", i, " name:", coins[i].name, ", id:", coins[i].id)
	// 	// db.Create(&Coin{Name: coins[i].name, CoinID: coins[i].id})
	// }

}

func enterCoinPrice() {
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("failed to connect database: ", err)
		os.Exit(1)
	}
	// db.AutoMigrate(&Price{})
	defer db.Close()
	var coins []Coin
	db.Find(&coins)
	// fmt.Println("db.Find(&coins):", result.RowsAffected, ", ", result.Error)
	// inputChannel := make(chan string)
	outputChannel := make(chan string)
	var res string
	j := 0
	for i := 0; i < len(coins); i++ {
		if coins[i].Watching == 1 {
			go getCoinInfo(coins[i].CoinID, outputChannel)
			j++
		}

		if j == numberOfRoutines {
			for t := 0; t < j; t++ {
				res = <-outputChannel
				if res == NoConnection {
					fmt.Println("No internet connectrion")
					return
				}
				// fmt.Println("res:", res)
				jsonparser.ArrayEach([]byte(res), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					id, _ := jsonparser.GetString(value, "id")
					temp, _ := jsonparser.GetString(value, "price")
					price, _ := strconv.ParseFloat(temp, 32)
					// fmt.Println(" price err:", err)
					temp, _ = jsonparser.GetString(value, "market_cap")
					marketCap, _ := strconv.ParseFloat(temp, 64)
					oneDayData, _, _, _ := jsonparser.Get(value, "1h")
					temp, _ = jsonparser.GetString([]byte(oneDayData), "volume")
					volume, _ := strconv.ParseFloat(temp, 64)
					fmt.Println(" price:", price, ", id:", id, ", market_cap:", marketCap, ", volume:", volume)
					db.Create(&Price{CoinID: id, Price: float32(price), TimeToSecond: time.Now().Unix(),
						MarketCap: uint64(marketCap), Volume: uint64(volume), SignalAI: 0, SignalAlg: 0})
				})
			}
			j = 0
		}
	}

	for t := 0; t < j; t++ {
		res = <-outputChannel
		if res == NoConnection {
			fmt.Println("No internet connectrion")
			return
		}
		// fmt.Println("res:", res)
		jsonparser.ArrayEach([]byte(res), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			id, _ := jsonparser.GetString(value, "id")
			temp, _ := jsonparser.GetString(value, "price")
			price, _ := strconv.ParseFloat(temp, 32)
			// fmt.Println(" price err:", err)
			temp, _ = jsonparser.GetString(value, "market_cap")
			marketCap, _ := strconv.ParseFloat(temp, 64)
			oneDayData, _, _, _ := jsonparser.Get(value, "1h")
			temp, _ = jsonparser.GetString([]byte(oneDayData), "volume")
			volume, _ := strconv.ParseFloat(temp, 64)
			fmt.Println(" price:", price, ", id:", id, ", market_cap:", marketCap, ", volume:", volume)
			db.Create(&Price{CoinID: id, Price: float32(price), TimeToSecond: time.Now().Unix(),
				MarketCap: uint64(marketCap), Volume: uint64(volume), SignalAI: 0, SignalAlg: 0})
		})
	}
	//
}

func getCoinInfo(coindID string, oChannel chan string) {
	resp, err := http.Get("https://api.nomics.com/v1/currencies/ticker?key=" + APIkey + "&ids=" + coindID + "&interval=1h&convert=USD")
	if err != nil {
		fmt.Println("failed to get coin info")
		// os.Exit(1)
		oChannel <- NoConnection
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("getCoinInfo failed to ioutil.ReadAll(resp.Body):", err)
		os.Exit(1)
	}
	data := string(body)
	oChannel <- data
}

func analysePrice() {
	fmt.Println("********** analysePrice *****************")
	db, err := gorm.Open(dbDriver, dbName)
	if err != nil {
		fmt.Println("failed to connect database: ", err)
		os.Exit(1)
	}
	defer db.Close()
	var coins []Coin
	db.Find(&coins)
	// result := db.Find(&coins)
	// fmt.Println("db.Find(&coins):", result.RowsAffected, ", ", result.Error)
	// inputChannel := make(chan string)
	// outputChannel := make(chan string)
	// var res string
	// j := 0
	for i := 0; i < len(coins); i++ {
		if coins[i].Watching == 1 {
			var prices []Price
			db.Where("coin_id = ? AND updated_at > ?", coins[i].CoinID, time.Now().AddDate(0, 0, -7)).Find(&prices)
			var sumVolume uint64
			var sumPrice float32
			sumVolume = 0
			sumPrice = 0.0
			for _, price := range prices {
				sumVolume = sumVolume + price.Volume
				// fmt.Println("CoinID:", price.CoinID, " , price:", price.Price,
				// 	", CreatedAt:", price.CreatedAt, ", MarketCap:", price.MarketCap, ", Volume:", price.Volume)
			}

			if len(prices) > 15 {
				for j := len(prices) - 1; j > len(prices)-16; j-- {
					// fmt.Println(prices[j].CoinID, " i:", j, " , len(prices):", len(prices))
					sumPrice = sumPrice + prices[j].Price
				}
				avgVolume := sumVolume / uint64(len(prices))
				avgPrice := sumPrice / 15.0
				lastPrice := prices[len(prices)-1]
				// secLastPrice := prices[len(prices)-2]
				fmt.Println("CoinID:", lastPrice.CoinID, ", avgPrice:", avgPrice, ", avgVolume:", avgVolume, " , price:", lastPrice.Price, ", Volume:", lastPrice.Volume)
				// fmt.Println("lastPrice CoinID:", lastPrice.CoinID, " , price:", lastPrice.Price,
				// 	", CreatedAt:", lastPrice.CreatedAt, ", MarketCap:", lastPrice.MarketCap, ", Volume:", lastPrice.Volume)
				// fmt.Println("secLastPrice CoinID:", secLastPrice.CoinID, " , price:", secLastPrice.Price,
				// 	", CreatedAt:", secLastPrice.CreatedAt, ", MarketCap:", secLastPrice.MarketCap, ", Volume:", secLastPrice.Volume)
				if lastPrice.Volume > uint64(float64(avgVolume)*1.10) && lastPrice.Price > (avgPrice*1.01) {
					var score float32
					score = float32(float32(lastPrice.Volume)/float32(avgVolume)) * float32(float32(lastPrice.Volume)/float32(avgVolume))
					fmt.Println("Bulish Signal score:", score, ", CoinID:", lastPrice.CoinID, " , price:", lastPrice.Price,
						", CreatedAt:", lastPrice.CreatedAt, ", MarketCap:", lastPrice.MarketCap, ", Volume:", lastPrice.Volume)
					db.Model(&lastPrice).Update("SignalAlg", 1)
				} else if lastPrice.Volume > uint64(float64(avgVolume)*1.10) && lastPrice.Price < (avgPrice*1.01) {
					var score float32
					score = float32(float32(lastPrice.Volume)/float32(avgVolume)) * float32(float32(lastPrice.Volume)/float32(avgVolume))
					fmt.Println("Bearish Signal score:", score, ", CoinID:", lastPrice.CoinID, " , price:", lastPrice.Price,
						", CreatedAt:", lastPrice.CreatedAt, ", MarketCap:", lastPrice.MarketCap, ", Volume:", lastPrice.Volume)
					db.Model(&lastPrice).Update("SignalAlg", -1)
				}
			}
		}
	}
}
