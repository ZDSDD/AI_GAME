package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Player struct {
	X, Y   int
	Health int
	Score  int
}

func NewPlayer(startPos [2]int) *Player {
	return &Player{X: startPos[0], Y: startPos[1], Health: 100, Score: 0}
}

func (p *Player) MoveTo(x, y int, dungeon *Dungeon) {
	// Ensure movement is only to adjacent tiles
	if abs(p.X-x)+abs(p.Y-y) != 1 {
		return
	}

	if x < 0 || y < 0 || x >= dungeonWidth || y >= dungeonHeight {
		return
	}

	cell := dungeon.Cells[y][x]

	switch cell.Type {
	case Empty, Entrance, Exit:
		p.X, p.Y = x, y
	case Monster:
		p.Health -= 10
		dungeon.Cells[y][x] = Cell{Type: Empty}
		p.Score += 10
	case Treasure:
		p.Score += 20
		dungeon.Cells[y][x] = Cell{Type: Empty}
	}
}

// Helper function to calculate absolute value
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (p *Player) Draw(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, float32(p.X*tileSize), float32(p.Y*tileSize), float32(tileSize), float32(tileSize), color.White, false)
}
