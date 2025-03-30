package main

import "github.com/hajimehoshi/ebiten/v2"

func HandleInput(g *Game) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		tileX, tileY := mouseX/tileSize, mouseY/tileSize

		if tileX < dungeonWidth && tileY < dungeonHeight {
			g.player.MoveTo(tileX, tileY, g.dungeon)
		}
	}
}
