package main

import (
	"fmt"

	"github.com/Speshl/gorrc_client/internal/app"
	"github.com/Speshl/gorrc_client/internal/config"
	socketio "github.com/googollee/go-socket.io"
)

func main() {
	cfg := config.GetConfig()

	socketURI := fmt.Sprintf("http://%s", cfg.Server)
	client, err := socketio.NewClient(socketURI, nil)
	if err != nil {
		err = fmt.Errorf("error creating client - %w", err)
		panic(err)
	}

	app := app.NewApp(cfg, client)

	app.RegisterHandlers()

	app.Start()
}
