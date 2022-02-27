package main

import (
	"github.com/AdamHyl/chatserver/common/log"
	"github.com/AdamHyl/chatserver/internal/conf"
	"github.com/AdamHyl/chatserver/internal/game"
)

func main() {
	logger, err := log.New(conf.Server.LogLevel, conf.Server.LogPath)
	if err != nil {
		log.Fatal("log init err")
		return
	}
	log.Export(logger)

	game.OnInit()
}
