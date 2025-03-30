package main

import (
	"fmt"

	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	dungeon        *Dungeon
	player         *Player
	hoverX, hoverY int
}

func NewGame() *Game {
	dungeon := NewDungeon(dungeonWidth, dungeonHeight)

	player := NewPlayer(dungeon.Entrance)

	return &Game{dungeon: dungeon, player: player}
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()
	g.hoverX, g.hoverY = mouseX/tileSize, mouseY/tileSize

	HandleInput(g, g.player)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.dungeon.Draw(screen, g.player)
	g.player.Draw(screen)

	// Highlight tile on hover
	if g.hoverX < dungeonWidth && g.hoverY < dungeonHeight {
		vector.DrawFilledRect(screen, float32(g.hoverX*tileSize), float32(g.hoverY*tileSize), float32(tileSize), float32(tileSize), color.RGBA{255, 255, 255, 100}, false)
	}

	// Display player stats
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Health: %d, Score: %d", g.player.Health, g.player.Score), 10, dungeonHeight*tileSize+10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
