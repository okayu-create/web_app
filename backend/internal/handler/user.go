package handler

import (
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/okayu-create/web_app/tree/main/backend/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// --- 1. 型定義（struct） ---

// 会員登録リクエストの構造体
type UserRegisterRequest struct {
	Name     string `json:"name" binding:"required"`        // binding:"required"で必須項目にする
	Email    string `json:"email" binding:"required,email"` // email形式もチェックする
	Password string `json:"password" binding:"required"`
}

// --- 2. ハンドラ定義 ---

// 会員登録を行う関数
func RegisterUserHandler(c *gin.Context) {
	var req UserRegisterRequest
	// HTTPリクエストボディをUserRegisterRequest構造体にバインドする
	// 必須項目や形式が異なる場合はエラーを返す
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HTTPリクエストボディのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "入力内容が正しくありません"})
		return
	}

	// メールアドレスの形式をチェックする
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9.]+@[a-zA-Z0-9.]+$`)
	if !emailPattern.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "正しいメールアドレスの形式で入力してください"})
		return
	}

	// パスワードの文字数をチェックする
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "パスワードは8文字以上で入力してください"})
		return
	}

	// メールアドレスの重複をチェックする
	db := database.GetDB()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&count)
	// QueryRow()メソッド自体のエラーもチェックする（DB接続エラーなど）
	if err != nil {
		log.Printf("メールアドレス重複チェックエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "このメールアドレスは既に登録されています"})
		return
	}

	// パスワードをハッシュ化する
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("パスワードハッシュ化エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// ユーザーをデータベースに登録する
	insertQuery := `
		INSERT INTO users (name, email, password, is_admin, enabled)
		VALUES (?, ?, ?, false, true)
	`
	_, err = db.Exec(insertQuery, req.Name, req.Email, string(hashedPassword))
	if err != nil {
		log.Printf("ユーザー登録エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// 登録成功のレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"message": "会員登録が完了しました"})
}
