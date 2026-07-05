package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okayu-create/web_app/tree/main/backend/internal/handler"
)

func main() {
	// Ginのデフォルトルータを作成
	router := gin.Default()

	// /api以下の各ルートを定義する
	api := router.Group("/api")
	{
		api.GET("/sample", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"message": "Hello, World!",
			})
		})
		api.GET("/products", handler.GetProductsHandler)
	}

	// Ginサーバをポート8080で起動する
	log.Println("Ginサーバをポート8080で起動します。")
	router.Run(":8080")
}
