package main

import (
	_ "goframe-starter/internal/packed"
	"goframe-starter/internal/service/bot"
)

func main() {
	bot.Job()
	//cmd.Main.Run(gctx.GetInitCtx())
}
