package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// This is now a global variable so it can be modified by game settings
var tileSize = 16

// Default constants that will be overridden by user settings
const (
	screenWidth  = 1280
	screenHeight = 720
)

func main() {

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Procedural Dungeon")

	// Create the main game with menu
	mainGame := NewMainGame()

	if err := ebiten.RunGame(mainGame); err != nil {
		log.Fatal(err)
	}
}
