package main

import (
	"embed"
	"log"
	"os"

	"passwordkeeper/internal/config"
	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/repo"
	"passwordkeeper/internal/service"
	"passwordkeeper/internal/transport"

	"github.com/BurntSushi/toml"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

type Config struct {
	DBPath string
}

var dbConfig Config

func main() {
	data, err := os.ReadFile("./config.toml")
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = toml.Decode(string(data), &dbConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := config.SQLiteService(dbConfig.DBPath)
	if err != nil {
		log.Fatal(err)
	}

	vaultRepo := repo.NewVaultRepository(db)
	itemrepo := repo.NewItemRepository(db)
	tagRepo := repo.NewTagRepository(db)
	itemTagRepo := repo.NewItemTagRepository(db)

	keyDeriver := crypto.NewArgon2KeyDeriver()
	encryptor := crypto.NewAESGCMEncryptor()
	session := service.NewMemoryVaultSession()

	vaultService := service.NewVaultService(vaultRepo, keyDeriver, encryptor, session)
	itemService := service.NewItemService(itemrepo, tagRepo, itemTagRepo, encryptor, session)

	handler := transport.NewHandler(vaultService, itemService)

	app := application.New(application.Options{
		Name:        "Password Keeper",
		Description: "An application for store your password",
		Services: []application.Service{
			application.NewService(handler),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Window 1",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
