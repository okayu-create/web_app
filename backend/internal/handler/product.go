package handler

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/okayu-create/web_app/tree/main/backend/internal/database"
)

// --- 1. 型定義（struct） ---
// フロントエンドが期待するJSONの「形」をGoの構造体で定義する
// `json:"..."`タグで、Goのフィールド名（大文字）とJSONのキー名（小文字）を対応させる

// 商品1つのデータを表す構造体
type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    int     `json:"price"`
	ImageURL *string `json:"image_url"` // NULL許容
}

// ページネーション情報を表す構造体
type Pagination struct {
	CurrentPage int `json:"currentPage"`
	PerPage     int `json:"perPage"`
	TotalItems  int `json:"totalItems"`
	TotalPages  int `json:"totalPages"`
}

// 商品一覧ページ全体のレスポンスを表す構造体
type ProductsPageData struct {
	Products   []Product  `json:"products"`
	Pagination Pagination `json:"pagination"`
}

// --- 2. ハンドラ定義 ---

// 商品一覧を返す関数
func GetProductsHandler(c *gin.Context) {
	// クエリパラメータ（?page=X）から「ページ番号」を取得する
	// クエリパラメータが存在しない、または不正な値の場合はデフォルトで1を設定する
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// クエリパラメータ（?perPage=X）から「1ページあたりの商品数」を取得する
	// クエリパラメータが存在しない、または不正な値の場合はデフォルトで16を設定する
	perPageStr := c.DefaultQuery("perPage", "16")
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 16
	}

	// データベース接続を取得する
	db := database.GetDB()

	// 商品の総数を取得するSQL文を実行する
	var totalItems int
	countQuery := "SELECT COUNT(*) FROM products"
	err = db.QueryRow(countQuery).Scan(&totalItems)
	if err != nil {
		log.Printf("商品総数の取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// ページネーション情報を計算する
	totalPages := int(math.Ceil(float64(totalItems) / float64(perPage)))
	offset := (page - 1) * perPage

	// 商品一覧を取得するSQL文を準備する
	query := `
			SELECT
				id,
				name,
				price,
				image_url
			FROM products
			LIMIT ? OFFSET ?
		`

	// SQL文を実行する
	rows, err := db.Query(query, perPage, offset)
	if err != nil {
		log.Printf("商品一覧の取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	defer rows.Close()

	// 商品データを格納するスライス
	products := []Product{}

	// SQL文の実行結果をスキャンする
	for rows.Next() {
		var p Product
		// image_urlカラムがNULLの場合に対応するため、sql.NullString型の変数imageUrlを用意する
		var imageUrl sql.NullString
		// 取得したデータをProduct構造体のフィールドにマッピングする
		// image_urlカラムは変数imageUrlにスキャンする
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &imageUrl); err != nil {
			log.Printf("商品データのスキャンエラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
			return
		}
		// 変数imageUrlがNULLでない場合にのみ、Product構造体のImageURLフィールドに値をセットする
		if imageUrl.Valid {
			p.ImageURL = &imageUrl.String
		}
		products = append(products, p)
	}
	if err = rows.Err(); err != nil {
		log.Printf("商品一覧データ取得中の行エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// ダミーのページネーション情報を作成する
	pagination := Pagination{
		CurrentPage: page,       // 現在のページ
		PerPage:     perPage,    // 1ページあたりの商品数
		TotalItems:  totalItems, // 全商品数
		TotalPages:  totalPages, // 総ページ数
	}

	// 最終的なレスポンスの形にまとめる
	response := ProductsPageData{
		Products:   products,
		Pagination: pagination,
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, response)
}
