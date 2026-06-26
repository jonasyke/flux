package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jonasyke/flux/internal"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx   context.Context
	db    *DBClient
	paths *internal.AppPaths
}

func NewApp(DBClient *DBClient, appPaths *internal.AppPaths) *App {
	return &App{
		db:    DBClient,
		paths: appPaths,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("Wails application started, database and paths are live.")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, falling back to system enviroment")
	}

	rootDir := internal.ResolveRootDir()
	appPaths := internal.NewAppPath(rootDir)

	if err := appPaths.EnsureDirsExist(); err != nil {
		log.Fatalf("Critical Error: FilePath could not be created: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	DBClient, err := NewDatabaseConnection(dbURL)
	if err != nil {
		log.Fatalf("Critical Error: Database initialization failed: %v", err)
	}

	defer DBClient.Close()

	app := NewApp(DBClient, &appPaths)

	err = wails.Run(&options.App{
		Title:  "Flux Mod Manager",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Wails execution encountered an error: %v", err)
	}
}
