package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Player struct {
	X, Y       int
	Health     int
	Score      int
	FOVEnabled bool
	FOVRadius  int
	Path       []Point // A list of points (tiles) the player will follow

}

func NewPlayer(startPos [2]int) *Player {
	return &Player{X: startPos[0], Y: startPos[1], Health: 100, Score: 0, FOVEnabled: false, FOVRadius: 6}
}

func (p *Player) MoveTo(targetX, targetY int, dungeon *Dungeon) {
	path := dungeon.FindPath(Point{p.X, p.Y}, Point{targetX, targetY})
	if len(path) > 1 {
		// Check if the next step is not a monster or treasure
		next := path[1]
		cell := dungeon.Cells[next.y][next.x]
		if cell.Type != Monster && cell.Type != Treasure {
			p.Path = path[1:] // Exclude current position
		}
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
func (p *Player) Update(dungeon *Dungeon) {
	if len(p.Path) > 0 {
		next := p.Path[0]
		cell := dungeon.Cells[next.y][next.x]

		// Stop if the next cell is not walkable
		if cell.Type == Monster || cell.Type == Treasure {
			p.Path = nil
			return
		}

		// Move to the next tile
		p.X, p.Y = next.x, next.y
		p.Path = p.Path[1:]
	}
}
