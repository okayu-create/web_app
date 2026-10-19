package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/okayu-create/web_app/tree/main/backend/internal/database"
)

// --- ハンドラ定義 ---

// 商品を新規登録する関数
func AdminCreateProductHandler(c *gin.Context) {
	// フォームデータを取得する
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	isFeaturedStr := c.PostForm("isFeatured")

	// 画像ファイルを取得する
	fileHeader, err := c.FormFile("imageFile")
	if err != nil {
		log.Printf("画像ファイルの取得エラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品画像は必須です"})
		return
	}

	// 入力データのバリデーションを行う
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品名は必須です"})
		return
	}
	price, err := strconv.Atoi(priceStr)
	if err != nil || price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "価格は0以上の整数で入力してください"})
		return
	}
	stock, err := strconv.Atoi(stockStr)
	if err != nil || stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "在庫数は0以上の整数で入力してください"})
		return
	}

	// チェックボックスがONであれば"on"、OFFであれば空文字が送信されてくる
	isFeatured := isFeaturedStr == "on"

	// descriptionが空であればデフォルト値を設定する
	if description == "" {
		description = "商品の説明がありません。"
	}

	// ファイル形式のバリデーションを行う
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "対応していないファイル形式です（jpg, jpeg, pngのみ）"})
		return
	}

	// ユニークなファイル名を生成する（タイムスタンプ + 乱数 + 拡張子）
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	random := rand.Intn(10000)
	fileName := fmt.Sprintf("%d_%d%s", timestamp, random, ext)

	// ファイルの保存先パス（DockerfileのWORKDIR［/app］からの相対パス）
	savePath := filepath.Join("uploads", fileName)

	// ファイルを保存する
	if err := c.SaveUploadedFile(fileHeader, savePath); err != nil {
		log.Printf("画像ファイルの保存エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ファイルのアップロードに失敗しました"})
		return
	}
	log.Printf("画像ファイルを保存しました: %s", savePath)

	// データーベースに商品情報を登録する
	db := database.GetDB()
	query := `
		INSERT INTO products (name, description, price, stock, image_url, is_featured)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = db.Exec(query, name, description, price, stock, fileName, isFeatured)
	if err != nil {
		log.Printf("商品登録エラー: %v", err)
		// 保存したファイルを削除する
		_ = os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// 登録成功のレスポンスを返す
	c.JSON(http.StatusCreated, gin.H{"message": "商品を登録しました"})
}
