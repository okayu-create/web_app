package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/okayu-create/web_app/tree/main/backend/internal/handler"
)

func main() {
	// Ginのデフォルトルータを作成
	router := gin.Default()

	// CORSを設定する
	router.Use(cors.New(cors.Config{
		// 許可するオリジン（フロントエンドが動いている場所）
		AllowOrigins: []string{"http://localhost:3000"},

		// 許可するHTTPメソッド
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		// 許可するHTTPヘッダー
		AllowHeaders: []string{"Content-Type"},

		// クッキーの送受信を許可する（認証で使うため）
		AllowCredentials: true,

		// プリフライトリクエストの結果をキャッシュする時間
		MaxAge: 12 * time.Hour,
	}))

	// /uploads/フォルダの内容を、URLパスの/uploads/で提供する
	// （DockerfileのWORKDIR［/app］から見た相対パス）
	router.StaticFS("/uploads", http.Dir("uploads"))

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
