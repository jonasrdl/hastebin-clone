package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jonasrdl/hastebin-clone/models"
	"math/big"
	"net/http"
	"time"
)

type PasteHandler struct {
	DB *sql.DB
}

func NewPasteHandler(db *sql.DB) *PasteHandler {
	return &PasteHandler{
		DB: db,
	}
}

func (h *PasteHandler) CreatePaste(c *gin.Context) {
	var paste models.Paste
	if err := c.BindJSON(&paste); err != nil {
		fmt.Printf("invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	password, err := generatePassword(16)
	if err != nil {
		fmt.Printf("error generating password: %v", err)
		return
	}

	paste.ID = uuid.New().String()
	paste.CreatedAt = time.Now()
	paste.Password = password

	// Insert the paste into the database
	_, err = h.DB.Exec(`
		INSERT INTO pastes (ID, Content, CreatedAt, Password) VALUES (?, ?, ?, ?)
	`, paste.ID, paste.Content, paste.CreatedAt, paste.Password)
	if err != nil {
		fmt.Printf("failed to create paste: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create paste"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": paste.ID, "password": paste.Password})
}

func (h *PasteHandler) GetPaste(c *gin.Context) {
	id := c.Param("id")
	var paste models.Paste

	// Retrieve the paste from the database
	err := h.DB.QueryRow(`
		SELECT ID, Content, CreatedAt, Password FROM pastes WHERE ID = ?
	`, id).Scan(&paste.ID, &paste.Content, &paste.CreatedAt, &paste.Password)
	if err == sql.ErrNoRows {
		fmt.Printf("paste not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Paste not found"})
		return
	} else if err != nil {
		fmt.Printf("failed to retrieve paste: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve paste"})
		return
	}

	// Check if Authorization header is present
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		// If not present, return a 401 Unauthorized with the static nonce
		nonce, err := generatePassword(10) // Use a different nonce for every request
		if err != nil {
			fmt.Printf("error generating nonce: %v", err)
			return
		}
		c.Header("WWW-Authenticate", `Basic realm="Paste Authorization", nonce="`+nonce+`"`)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, paste.Content)
}

// generatePassword generates a random password of the specified length
func generatePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:'\",.<>/?`~"

	seed := make([]byte, 64)
	_, err := rand.Read(seed)
	if err != nil {
		return "", err
	}

	password := make([]byte, length)
	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[randomIndex.Int64()]
	}

	return string(password), nil
}
