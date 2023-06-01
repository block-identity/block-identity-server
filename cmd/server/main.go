package main

import (
	"database/sql"
	"github.com/block-identity/block-identity-server/api"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/block-identity/block-identity-server/pkg"
)

func initDB() error {
	var err error
	pkg.DB, err = sql.Open("sqlite3", "./user.db")
	if err != nil {
		return err
	}

	_, err = pkg.DB.Exec("CREATE TABLE IF NOT EXISTS users (id TEXT, platform TEXT, secret_key TEXT)")
	if err != nil {
		return err
	}
	return nil
}

func init() {
	pkg.ClientID = os.Getenv("CLIENT_ID")
	if pkg.ClientID == "" {
		panic("CLIENT_ID not provided")
	}
	pkg.ClientSecret = os.Getenv("CLIENT_SECRET")
	if pkg.ClientSecret == "" {
		panic("CLIENT_SECRET not provided")
	}
	pkg.Oauth2Config = &oauth2.Config{
		ClientID:     pkg.ClientID,
		ClientSecret: pkg.ClientSecret,
		RedirectURL:  pkg.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}
	err := initDB()
	if err != nil {
		panic("failed to connect database")
	}
}

func main() {
	router := api.GetRouter()
	router.Run()
	pkg.DB.Close()
}
