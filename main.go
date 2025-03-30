package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth   = 640
	screenHeight  = 480
	tileSize      = 32
	dungeonWidth  = 20
	dungeonHeight = 15
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Procedural Dungeon")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
