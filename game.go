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

// You'll also need to adjust the Update method to account for the margins when calculating hover position

func (g *Game) Update() error {
	// Define the same margin values used in Draw
	const marginX, marginY = 20, 40

	mouseX, mouseY := ebiten.CursorPosition()

	// Adjust mouse coordinates to account for margins
	adjustedMouseX := mouseX - marginX
	adjustedMouseY := mouseY - marginY

	// Convert to tile coordinates (if within the valid area)
	if adjustedMouseX >= 0 && adjustedMouseY >= 0 {
		g.hoverX, g.hoverY = adjustedMouseX/tileSize, adjustedMouseY/tileSize
	} else {
		// Mouse is in the margin area
		g.hoverX, g.hoverY = -1, -1
	}

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

	// Update the message timestamps
	g.interactionHandler.UpdateMessages()

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
	// Define margin values
	const marginX, marginY = 20, 40 // You can adjust these values as needed

	// Create a rendering context with translation for the margins
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(marginX), float64(marginY))

	// Use a sub-screen approach to implement the margin
	dungeonScreen := ebiten.NewImage(screenWidth-2*marginX, screenHeight-2*marginY)

	// Draw dungeon to the sub-screen
	g.dungeon.Draw(dungeonScreen, g.player)

	// Draw path to hover before drawing the player
	if len(g.pathToHover) > 0 {
		for i, p := range g.pathToHover {
			// Reverse the gradient calculation, the closer tiles are more visible
			gradient := float32(len(g.pathToHover)-i) / float32(len(g.pathToHover)) // Closer tiles have higher gradient
			shade := uint8(60 + 40*gradient)
			// Gradually fade the alpha as the path goes further
			alpha := uint8(120 - 50*gradient)

			pathColor := color.RGBA{
				shade,
				shade,
				uint8(shade + 10),
				alpha,
			}

			vector.DrawFilledRect(
				dungeonScreen,
				float32(p[0]*tileSize),
				float32(p[1]*tileSize),
				float32(tileSize),
				float32(tileSize),
				pathColor,
				false,
			)
		}
	}

	// Draw player on the sub-screen
	g.player.Draw(dungeonScreen)

	// Draw the sub-screen to the main screen with margins
	screen.DrawImage(dungeonScreen, op)

	// Highlight the hovered tile (needs to be adjusted for margins)
	if g.hoverX < g.dungeon.Width && g.hoverY < g.dungeon.Height {
		vector.StrokeRect(
			screen,
			float32(g.hoverX*tileSize+marginX),
			float32(g.hoverY*tileSize+marginY),
			float32(tileSize),
			float32(tileSize),
			1.5, // thickness
			color.RGBA{255, 255, 255, 180},
			false,
		)

		// Show info about the hovered cell (adjusted for margins)
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

			ebitenutil.DebugPrintAt(screen, cellInfo, g.hoverX*tileSize+marginX, g.hoverY*tileSize+marginY-10)
		}
	}

	// Display player stats (at the top with some padding)
	statY := 10
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Health: %d/%d, Score: %d | Dungeon Level: %d",
		g.player.Health, g.player.MaxHealth, g.player.Score, g.dungeon.Level), 10, statY)
	statY += 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Player Level: %d | Defense: %d | Luck: %d",
		g.player.Level, g.player.Defense, g.player.Luck), 10, statY)

	// Display interaction messages with very subtle transparency
	messages := g.interactionHandler.GetMessages()
	if len(messages) > 0 {
		// No background box - keep it minimal
		statY += 15

		// Use very faint text for all messages
		for i, msg := range messages {
			// Calculate alpha value based on message age - make ALL messages very subtle
			// Starting with a very low base alpha
			baseAlpha := 100 // Much lower base alpha
			alpha := uint8(baseAlpha - (i * 20))
			if alpha < 25 {
				alpha = 25 // Minimum visibility
			}

			// Draw a very subtle background for each message
			vector.DrawFilledRect(
				screen,
				10,
				float32(statY-2),
				300,
				16,
				color.RGBA{0, 0, 0, alpha / 3}, // Very low alpha for the background
				false,
			)

			// Draw the message text
			// Using a lower alpha value for the background
			// Note: We can't directly control text alpha with DebugPrintAt

			// Use a short prefix for less visual impact
			ebitenutil.DebugPrintAt(
				screen,
				fmt.Sprintf("· %s", msg), // Smaller bullet point
				12,
				statY)
			statY += 15 // Reduced line spacing
			statY += 20
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
