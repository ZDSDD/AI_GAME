package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	tileSize     = 16
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Procedural Dungeon")

	// Generate random dungeon dimensions
	randomWidth := 40 + rand.Intn(30) // 40–69 tiles wide
	randomHeight := 12 + rand.Intn(8) // 12–19 tiles high

	game := NewGame(randomWidth, randomHeight)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
