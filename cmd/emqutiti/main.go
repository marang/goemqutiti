package main

import (
	app "github.com/marang/emqutiti"
	cfg "github.com/marang/emqutiti/cmd"
)

func main() {
	app.Main(cfg.ParseFlags())
}
