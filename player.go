package main

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Player struct {
	X, Y         int
	Health       int
	MaxHealth    int
	Score        int
	FOVEnabled   bool
	FOVRadius    int
	moveCooldown int // frames until next move

	Path []Point // A list of points (tiles) the player will follow

	// New player stats that affect interactions
	Defense    int // Reduces damage from monsters
	Luck       int // Increases treasure value
	Level      int // Player's current level
	Experience int // Experience points
}

func NewPlayer(startPos [2]int) *Player {
	return &Player{
		X:            startPos[0],
		Y:            startPos[1],
		Health:       100,
		MaxHealth:    100,
		Score:        0,
		FOVEnabled:   true,
		FOVRadius:    6,
		Defense:      10, // 10% damage reduction
		Luck:         5,  // 5% treasure value increase
		Level:        1,
		Experience:   0,
		moveCooldown: 0,
	}
}

func (p *Player) MoveTo(targetX, targetY int, dungeon *Dungeon, interactionHandler *InteractionHandler) {
	path := dungeon.FindPath(Point{p.X, p.Y}, Point{targetX, targetY})
	if len(path) > 1 {
		next := path[1]
		cell := dungeon.Cells[next.y][next.x]

		// Handle interaction for special cells
		if cell.Type == Monster || cell.Type == Treasure || cell.Type == Exit {
			result := interactionHandler.Handle(cell.Type, p)

			// If the interaction removes the entity, clear the cell
			if result.RemoveEntity {
				dungeon.Cells[next.y][next.x].Type = Empty
			}

			// Special handling for exit
			if cell.Type == Exit {
				// Generate new random dimensions for the next dungeon
				newWidth := 40 + rand.Intn(30) // 40–69
				newHeight := 12 + rand.Intn(8) // 12–19

				newLevel := dungeon.Level + 1
				*dungeon = *NewDungeon(newWidth, newHeight, newLevel)

				// Move player to the new entrance
				p.X, p.Y = dungeon.Entrance[0], dungeon.Entrance[1]
				return
			}

			// Move to the cell if it's now empty
			if dungeon.Cells[next.y][next.x].Type == Empty {
				p.Path = path[1:2] // Just move one step
			}
		} else {
			// Normal movement for empty cells
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
	if p.moveCooldown > 0 {
		p.moveCooldown--
		return
	}

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

		// Reset movement delay (e.g., 10 frames)
		p.moveCooldown = 10
	}
}
