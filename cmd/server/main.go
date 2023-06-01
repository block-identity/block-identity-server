package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID        string `json:"id"`
	Platform  string `json:"platform"`
	SecretKey string `json:"secret_key"`
}

var (
	clientID     string
	clientSecret string
	redirectURL  = "http://localhost:8080/auth"
	DB           *sql.DB

	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}
)

func initDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./user.db")
	if err != nil {
		return err
	}

	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS users (id TEXT, platform TEXT, secret_key TEXT)")
	if err != nil {
		return err
	}
	return nil
}

func init() {
	clientID = os.Getenv("CLIENT_ID")
	if clientID == "" {
		panic("CLIENT_ID not provided")
	}
	clientSecret = os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		panic("CLIENT_SECRET not provided")
	}
	err := initDB()
	if err != nil {
		panic("failed to connect database")
	}
}

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "welcome to block identity server",
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/auth", handleAuth)
	r.POST("/test", testAuth)
	r.Run()
	DB.Close()
}

func handleAuth(c *gin.Context) {
	ctx := context.Background()
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not provided"})
		return
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !token.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not valid"})
		return
	}
	fmt.Println("accessToken:", token.AccessToken)

	service, err := v2.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userinfo, err := service.Userinfo.Get().Do()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("id:", userinfo.Id)
	fmt.Println("userinfo:", userinfo)

	c.JSON(http.StatusOK, gin.H{"message": "User authenticated successfully"})
}

func generateEthereumSecretKey() string {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	return fmt.Sprintf("%x", privateKeyBytes)
}

func testAuth(c *gin.Context) {
	fmt.Println("credential パラメータを表示します。")
	fmt.Println(c.GetQuery("credential"))
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
