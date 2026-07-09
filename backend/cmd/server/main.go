package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/okayu-create/web_app/tree/main/backend/internal/handler"
)

func main() {
	// マイグレーションを実行する
	runMigration()

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

		// preflightリクエストの結果をキャッシュする時間
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

// マイグレーションを実行する関数
func runMigration() {
	// 環境変数からDSN（データソース名）を取得する
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("エラー：環境変数DB_DSNが設定されていません")
	}

	log.Println("マイグレーションを開始します...")

	// データベースへの接続準備を行う
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("データベースへの接続準備に失敗しました: %v", err)
	}

	// マイグレーションが終わったらデータベース接続を閉じる
	defer db.Close()

	// データベースへの接続を試みる
	err = db.Ping()
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}

	// マイグレーション用のデータベースドライバーを作成する
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("マイグレーション用のデータベースドライバーの作成に失敗しました：%v", err)
	}

	// マイグレーションのソース（SQLファイル）の場所を指定する
	// Dockerfile.devファイルのWORKDIR（/app）からの相対パス
	sourceURL := "file://db/migrations"

	// マイグレーション用のインスタンスを作成する
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "mysql", driver)
	if err != nil {
		log.Fatalf("マイグレーションのセットアップに失敗しました：%v", err)
	}

	// マイグレーションを実行する
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("マイグレーションに失敗しました：%v", err)
	}

	log.Println("マイグレーションに成功しました！")
}
