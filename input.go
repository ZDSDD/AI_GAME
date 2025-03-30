package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

var prevKeyState bool

func HandleInput(g *Game, player *Player) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		tileX, tileY := mouseX/tileSize, mouseY/tileSize

		if tileX < dungeonWidth && tileY < dungeonHeight {
			g.player.MoveTo(tileX, tileY, g.dungeon)
		}
	}
	keyPressed := ebiten.IsKeyPressed(ebiten.KeyF)

	// Toggle only when transitioning from released -> pressed
	if keyPressed && !prevKeyState {
		player.FOVEnabled = !player.FOVEnabled
	}

	// Store current key state for next frame
	prevKeyState = keyPressed
}
