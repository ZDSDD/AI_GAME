package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

var prevKeyState bool

// Handle player input and toggle FOV
func HandleInput(g *Game, player *Player) {
	// Handle mouse input for movement
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		tileX, tileY := mouseX/tileSize, mouseY/tileSize

		if tileX < dungeonWidth && tileY < dungeonHeight {
			// Move player to the tile clicked on
			g.player.MoveTo(tileX, tileY, g.dungeon)
		}
	}

	// Handle keyboard input for toggling FOV
	keyPressed := ebiten.IsKeyPressed(ebiten.KeyF)

	// Toggle FOV only when transitioning from released -> pressed
	if keyPressed && !prevKeyState {
		player.FOVEnabled = !player.FOVEnabled
	}

	// Store current key state for next frame
	prevKeyState = keyPressed
}
