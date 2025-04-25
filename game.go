package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	dungeon            *Dungeon
	player             *Player
	hoverX, hoverY     int
	pathToHover        [][2]int
	interactionHandler *InteractionHandler
}

func NewGame(width, height int) *Game {
	dungeon := NewDungeon(width, height, 1)
	player := NewPlayer(dungeon.Entrance)

	// Create the interaction handler
	interactionHandler := NewInteractionHandler()

	// Register interactions for different cell types
	interactionHandler.Register(Monster, NewMonsterInteraction(1))            // Default level 1
	interactionHandler.Register(Treasure, NewTreasureInteraction(10, "gold")) // Default 10 gold
	interactionHandler.Register(Exit, NewExitInteraction(2))                  // Go to level 2

	return &Game{
		dungeon:            dungeon,
		player:             player,
		interactionHandler: interactionHandler,
	}
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()
	g.hoverX, g.hoverY = mouseX/tileSize, mouseY/tileSize

	// Calculate path to hover position
	if g.hoverX >= 0 && g.hoverX < g.dungeon.Width && g.hoverY >= 0 && g.hoverY < g.dungeon.Height {
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

	// Update interaction logic for cell types that change each level
	// This ensures that when a new level is generated, the interaction
	// system has the correct values for each cell type
	for y := 0; y < g.dungeon.Height; y++ {
		for x := 0; x < g.dungeon.Width; x++ {
			cell := g.dungeon.Cells[y][x]

			switch cell.Type {
			case Monster:
				g.interactionHandler.Register(Monster, NewMonsterInteraction(cell.InteractionLevel))
			case Treasure:
				g.interactionHandler.Register(Treasure, NewTreasureInteraction(cell.InteractionLevel, cell.TreasureType))
			case Exit:
				g.interactionHandler.Register(Exit, NewExitInteraction(g.dungeon.Level+1))
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.dungeon.Draw(screen, g.player)

	// Draw path to hover before drawing the player
	if len(g.pathToHover) > 0 {
		for i, p := range g.pathToHover {
			// Use a gradient from blue to red based on distance
			// Use a subtle gradient of cool grays
			gradient := float32(i) / float32(len(g.pathToHover))
			shade := uint8(60 + 40*gradient) // Range: 60–100
			pathColor := color.RGBA{
				shade,
				shade,
				uint8(shade + 10), // Slight bluish tint
				70,                // More subtle transparency
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
		}
	}

	g.player.Draw(screen)

	// Highlight the hovered tile
	if g.hoverX < g.dungeon.Width && g.hoverY < g.dungeon.Height {
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

		// Show info about the hovered cell
		if g.hoverX >= 0 && g.hoverY >= 0 && g.hoverX < g.dungeon.Width && g.hoverY < g.dungeon.Height {
			cell := g.dungeon.Cells[g.hoverY][g.hoverX]
			var cellInfo string

			switch cell.Type {
			case Monster:
				cellInfo = fmt.Sprintf("Monster (Level %d)", cell.InteractionLevel)
			case Treasure:
				cellInfo = fmt.Sprintf("%s (Value %d)", cell.TreasureType, cell.InteractionLevel)
			case Exit:
				cellInfo = fmt.Sprintf("Exit to Level %d", cell.InteractionLevel)
			case Entrance:
				cellInfo = "Entrance"
			case Empty:
				cellInfo = "Empty"
			case Wall:
				cellInfo = "Wall"
			}

			ebitenutil.DebugPrintAt(screen, cellInfo, g.hoverX*tileSize, g.hoverY*tileSize-10)
		}
	}

	// Display player stats
	statY := 10
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Health: %d/%d, Score: %d | Dungeon Level: %d",
		g.player.Health, g.player.MaxHealth, g.player.Score, g.dungeon.Level), 10, statY)
	statY += 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Player Level: %d | Defense: %d | Luck: %d",
		g.player.Level, g.player.Defense, g.player.Luck), 10, statY)

	// Display interaction messages
	statY += 30
	messages := g.interactionHandler.GetMessages()
	if len(messages) > 0 {
		ebitenutil.DebugPrintAt(screen, "Recent events:", 10, statY)
		statY += 15
		for i, msg := range messages {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d. %s", i+1, msg), 10, statY)
			statY += 15
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
