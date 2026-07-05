package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// --- 1. 型定義（struct） ---
// フロントエンドが期待するJSONの「形」をGoの構造体で定義する
// `json:"..."`タグで、Goのフィールド名（大文字）とJSONのキー名（小文字）を対応させる

// 商品1つのデータを表す構造体
type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	ImageURL string `json:"image_url"`
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
	// ダミーの商品データを3つ作成する
	products := []Product{
		{
			ID:       1,
			Name:     "イヤホン",
			Price:    5000,
			ImageURL: "product01.jpg",
		},
		{
			ID:       2,
			Name:     "ブランケット",
			Price:    3000,
			ImageURL: "product02.jpg",
		},
		{
			ID:       3,
			Name:     "折りたたみ傘",
			Price:    1500,
			ImageURL: "product03.jpg",
		},
	}

	// ダミーのページネーション情報を作成する
	pagination := Pagination{
		CurrentPage: 1,  // 現在のページ
		PerPage:     16, // 1ページあたりの商品数
		TotalItems:  3,  // 全商品数
		TotalPages:  1,  // 総ページ数
	}

	// 最終的なレスポンスの形にまとめる
	response := ProductsPageData{
		Products:   products,
		Pagination: pagination,
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, response)
}
