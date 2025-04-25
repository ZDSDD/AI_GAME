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

	// Calculate path to hover position
	if g.hoverX >= 0 && g.hoverX < dungeonWidth && g.hoverY >= 0 && g.hoverY < dungeonHeight {
		// Get the path from player position to hover position
		path := g.dungeon.FindPath(Point{g.player.X, g.player.Y}, Point{g.hoverX, g.hoverY})

		// Convert path to [][2]int format for rendering
		g.pathToHover = nil
		if path != nil {
			for i := 1; i < len(path); i++ { // Skip the first point (player's position)
				point := path[i]
				// Check if we should stop at this point (monster or treasure)
				if g.dungeon.Cells[point.y][point.x].Type == Monster ||
					g.dungeon.Cells[point.y][point.x].Type == Treasure {
					// Add this point to the path (so it's highlighted)
					g.pathToHover = append(g.pathToHover, [2]int{point.x, point.y})
					break
				}
				g.pathToHover = append(g.pathToHover, [2]int{point.x, point.y})
			}
		}
	} else {
		g.pathToHover = nil
	}

	HandleInput(g, g.player)
	g.player.Update(g.dungeon)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.dungeon.Draw(screen, g.player)

	// Draw path to hover before drawing the player
	if len(g.pathToHover) > 0 {
		for i, p := range g.pathToHover {
			// Use a gradient from blue to red based on distance
			gradient := float32(i) / float32(len(g.pathToHover))
			pathColor := color.RGBA{
				uint8(50 + 150*gradient),  // Increase red as we get further
				50,                        // Constant green
				uint8(200 - 150*gradient), // Decrease blue as we get further
				100,                       // Semi-transparent
			}

			// Draw the path tile
			vector.DrawFilledRect(
				screen,
				float32(p[0]*tileSize),
				float32(p[1]*tileSize),
				float32(tileSize),
				float32(tileSize),
				pathColor,
				false,
			)

			// Add a small border to make it more visible
			vector.StrokeRect(
				screen,
				float32(p[0]*tileSize),
				float32(p[1]*tileSize),
				float32(tileSize),
				float32(tileSize),
				1.0,                            // Thin border
				color.RGBA{255, 255, 255, 200}, // White border
				false,
			)
		}
	}

	g.player.Draw(screen)

	// Highlight the hovered tile
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
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
