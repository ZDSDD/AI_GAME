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
	pathToHover    [][2]int
}

func NewGame() *Game {
	dungeon := NewDungeon(dungeonWidth, dungeonHeight, 1)
	player := NewPlayer(dungeon.Entrance)
	return &Game{dungeon: dungeon, player: player}
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()
	g.hoverX, g.hoverY = mouseX/tileSize, mouseY/tileSize

	HandleInput(g, g.player)
	g.player.Update(g.dungeon) // Animate player path

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.dungeon.Draw(screen, g.player)
	g.player.Draw(screen)

	// Highlight tile on hover
	if g.hoverX < dungeonWidth && g.hoverY < dungeonHeight {
		vector.StrokeRect(
			screen,
			float32(g.hoverX*tileSize),
			float32(g.hoverY*tileSize),
			float32(tileSize),
			float32(tileSize),
			1.5, // thickness
			color.RGBA{255, 255, 255, 180},
			false,
		)

	}

	// Display player stats
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Health: %d, Score: %d | Dungeon Level: %d", g.player.Health, g.player.Score, g.dungeon.Level), 10, 10)
	for _, p := range g.pathToHover {
		vector.DrawFilledRect(screen, float32(p[0]*tileSize), float32(p[1]*tileSize), float32(tileSize), float32(tileSize), color.RGBA{20, 30, 50, 50}, false)
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
