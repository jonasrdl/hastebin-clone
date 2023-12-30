package handlers

import (
	"crypto/rand"
	"database/sql"
	"errors"
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
	DB     *sql.DB
	APIKey string
}

func NewPasteHandler(db *sql.DB, apiKey string) *PasteHandler {
	return &PasteHandler{
		DB:     db,
		APIKey: apiKey,
	}
}

func (h *PasteHandler) CreatePaste(c *gin.Context) {
	// Check if Authorization header is present and matches the API key
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key is required"})
		return
	}

	if authHeader != h.APIKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
		return
	}

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

	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		fmt.Printf("error loading CET timezone: %v", err)
		return
	}

	paste.ID = uuid.New().String()
	paste.CreatedAt = time.Now().In(location)
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
	if errors.Is(err, sql.ErrNoRows) {
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
		// If not present, check for the password query parameter
		queryPassword := c.Request.URL.Query().Get("password")
		if queryPassword == "" || queryPassword != paste.Password {
			// If no password query parameter is provided, or it doesn't match the paste password, return a 401 Unauthorized
			c.Header("WWW-Authenticate", `Basic realm="Paste Authorization"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	} else {
		// If Authorization header is present, check for basic auth
		_, password, ok := c.Request.BasicAuth()
		if !ok || password != paste.Password {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
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
