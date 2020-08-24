package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	initalizeRoutes()
}

func initalizeRoutes() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/css", "templates/css")
	router.Static("/js", "templates/js")

	router.GET("/", showIndexPage)
	router.GET("/coin-price/:coinID", routeGetCoinPrice)
	router.GET("/watching-coin", routeGetWatchingCoins)

	router.Run(":8085") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func showIndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Main website",
	})
}

func routeGetCoinPrice(c *gin.Context) {
	coinID := c.Param("coinID")
	var prices []Price
	prices = getCoinPrice(coinID)
	res, _ := json.Marshal(prices)
	c.String(http.StatusOK, string(res))
}

func routeGetWatchingCoins(c *gin.Context) {
	var coins []Coin
	coins = getWatchingCoins()
	res, _ := json.Marshal(coins)
	c.String(http.StatusOK, string(res))
}
