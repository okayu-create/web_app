package handler

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okayu-create/web_app/tree/main/backend/internal/database"
)

// --- 1. 型定義（struct） ---
// フロントエンドが期待するJSONの「形」をGoの構造体で定義する
// `json:"..."`タグで、Goのフィールド名（大文字）とJSONのキー名（小文字）を対応させる

// 商品一覧用の構造体
type ProductListItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    int     `json:"price"`
	ImageURL *string `json:"image_url"` // NULL許容
}

// 商品詳細用の構造体
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"` // NULL許容
	Price       int       `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    *string   `json:"image_url"` // NULL許容
	SalesCount  int       `json:"sales_count"`
	IsFeatured  bool      `json:"is_featured"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	Products   []ProductListItem `json:"products"`
	Pagination Pagination        `json:"pagination"`
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

	// クエリパラメータ(?sort=x)から「並べ替え条件」を取得する（デフォルトは新着順）
	sort := c.DefaultQuery("sort", "new")

	// 並べ替え条件（ORDER BY句）を構築する
	var orderByClause string
	switch sort {
	case "priceAsc": //	価格が安い順
		orderByClause = "ORDER BY price ASC"
	case "new": // 新着順
		orderByClause = "ORDER BY created_at DESC"
	default:
		// 上記条件以外はデフォルトの新着順にする
		orderByClause = "ORDER BY created_at DESC"
	}

	// クエリパラメータ（?keyword=x）から「検索キーワード」を取得する
	keyword := c.DefaultQuery("keyword", "")

	// 検索条件（WHERE句）を構築する
	var whereClause string
	var whereParams []any // WHERE句用のパラメータを格納するスライス
	if keyword != "" {
		whereClause = "WHERE (name LIKE ? OR description LIKE ?)"
		// パラメータ（%キーワード%）をスライスに格納する
		likekeyword := "%" + keyword + "%"
		whereParams = append(whereParams, likekeyword, likekeyword)
	}

	// データベース接続を取得する
	db := database.GetDB()

	// 商品の総数を取得するSQL文を実行する
	var totalItems int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	// whereParams...でスライスの内容を展開してQueryRow()メソッドに渡す
	err = db.QueryRow(countQuery, whereParams...).Scan(&totalItems)
	if err != nil {
		log.Printf("商品総数の取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// ページネーション情報を計算する
	totalPages := int(math.Ceil(float64(totalItems) / float64(perPage)))
	offset := (page - 1) * perPage

	// 商品一覧を取得するSQL文を準備する
	query := fmt.Sprintf(`
			SELECT
				id,
				name,
				price,
				image_url
			FROM products
			%s
			%s
			LIMIT ? OFFSET ?
		`, whereClause, orderByClause) // %sの部分に変数whereClauseとorderByClauseが挿入される

	// SQL文に渡すパラメータを準備する
	queryParams := append(whereParams, perPage, offset)

	// SQL文を実行する
	rows, err := db.Query(query, queryParams...)
	if err != nil {
		log.Printf("商品一覧の取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	defer rows.Close()

	// 商品データを格納するスライス
	products := []ProductListItem{}

	// SQL文の実行結果をスキャンする
	for rows.Next() {
		var p ProductListItem
		// image_urlカラムがNULLの場合に対応するため、sql.NullString型の変数imageUrlを用意する
		var imageUrl sql.NullString
		// 取得したデータをProductListItem構造体のフィールドにマッピングする
		// image_urlカラムは変数imageUrlにスキャンする
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &imageUrl); err != nil {
			log.Printf("商品データのスキャンエラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
			return
		}
		// 変数imageUrlがNULLでない場合にのみ、ProductListItem構造体のImageURLフィールドに値をセットする
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

// 商品詳細を返す関数
func GetProductByIDHandler(c *gin.Context) {
	// パスパラメータ（.../product/:idのうち、:idの部分）から「商品ID」を取得する
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("不正な商品ID形式です：%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品IDが不正です"})
		return
	}

	// データベース接続を取得する
	db := database.GetDB()

	// 商品詳細を取得するSQL文を準備する
	query := `
		SELECT
			id, name,
			description,
			price,
			stock,
			image_url,
			sales_count,
			is_featured,
			created_at,
			updated_at
		FROM products
		WHERE id = ?
	`
	var p Product
	var description sql.NullString
	var imageUrl sql.NullString

	// SQL文を実行して結果をスキャンする
	err = db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Name,
		&description,
		&p.Price,
		&p.Stock,
		&imageUrl,
		&p.SalesCount,
		&p.IsFeatured,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("商品が見つかりません：ID=%d", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "商品が見つかりませんでした"})
		} else {
			log.Printf("商品取得エラー（ID=%d）：%v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		}
		return
	}

	// 変数descriptionと変数imageUrlがNULLでない場合にのみ、Product構造体の対応するフィールドに値をセットする
	if description.Valid {
		p.Description = &description.String
	}
	if imageUrl.Valid {
		p.ImageURL = &imageUrl.String
	}

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, p)
}
