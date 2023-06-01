package handler

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/block-identity/block-identity-server/pkg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"log"
	"net/http"
)

func Auth(c *gin.Context) {
	ctx := context.Background()
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not provided"})
		return
	}

	token, err := pkg.Oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !token.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not valid"})
		return
	}
	fmt.Println("accessToken:", token.AccessToken)

	service, err := v2.NewService(ctx, option.WithTokenSource(pkg.Oauth2Config.TokenSource(ctx, token)))
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

	var user pkg.User
	err = pkg.DB.QueryRow("SELECT * FROM users WHERE id = ?", userinfo.Id).Scan(&user.ID, &user.Platform, &user.SecretKey)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("新規ユーザを作成します")
			// create new user
			user = pkg.User{
				ID:        userinfo.Id,
				Platform:  "google",
				SecretKey: generateEthereumSecretKey(),
			}
			_, err = pkg.DB.Exec("INSERT INTO users (id, platform, secret_key) VALUES (?, ?, ?)", user.ID, user.Platform, user.SecretKey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  user.ID,
		"platform": user.Platform,
	})
}

func generateEthereumSecretKey() string {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	return fmt.Sprintf("%x", privateKeyBytes)
}
