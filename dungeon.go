package main

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type CellType int

const (
	Empty CellType = iota
	Wall
	Monster
	Treasure
	Entrance
	Exit
)

type Cell struct {
	Type CellType
}

type Dungeon struct {
	Cells         [][]Cell
	Width, Height int
	Entrance      [2]int
	Exit          [2]int
}

const (
	FOVRadius    = 5
	NumMonsters  = 5
	NumTreasures = 5
)

func NewDungeon(width, height int) *Dungeon {

	d := &Dungeon{
		Cells:  make([][]Cell, height),
		Width:  width,
		Height: height,
	}

	// Initialize all cells as walls
	for y := 0; y < height; y++ {
		d.Cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			d.Cells[y][x] = Cell{Type: Wall}
		}
	}

	// Generate maze with proper paths
	d.generateMaze()

	// Place entrance
	entranceX, entranceY := d.placeRandomFeature(Empty, Entrance)
	d.Entrance = [2]int{entranceX, entranceY}

	// Place exit (ensure it's far from entrance)
	var exitX, exitY int
	for {
		exitX, exitY = d.placeRandomFeature(Empty, Exit)
		// Check if exit is at least 1/3 of the dungeon size away from the entrance
		minDistance := (width + height) / 3
		dx, dy := entranceX-exitX, entranceY-exitY
		distance := dx*dx + dy*dy
		if distance >= minDistance*minDistance {
			d.Exit = [2]int{exitX, exitY}
			break
		}
		// Revert and try again
		d.Cells[exitY][exitX] = Cell{Type: Empty}
	}

	// Place monsters in valid locations (empty cells only)
	for i := 0; i < NumMonsters; i++ {
		d.placeRandomFeature(Empty, Monster)
	}

	// Place treasures in valid locations (empty cells only)
	for i := 0; i < NumTreasures; i++ {
		d.placeRandomFeature(Empty, Treasure)
	}

	return d
}

// Helper function to place a feature in a random empty cell
func (d *Dungeon) placeRandomFeature(requiredType, newType CellType) (int, int) {
	for {
		x, y := rand.Intn(d.Width-2)+1, rand.Intn(d.Height-2)+1
		if d.Cells[y][x].Type == requiredType {
			d.Cells[y][x] = Cell{Type: newType}
			return x, y
		}
	}
}

// Properly implemented Prim's algorithm for maze generation
func (d *Dungeon) generateMaze() {
	// Start with a grid full of walls
	width, height := d.Width, d.Height

	// Create a list to track cells that have been visited
	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	// Pick a random starting point
	startX, startY := rand.Intn(width-2)+1, rand.Intn(height-2)+1
	d.Cells[startY][startX] = Cell{Type: Empty}
	visited[startY][startX] = true

	// Add walls of the starting cell to the wall list
	walls := []struct{ x, y int }{}

	// Add neighboring walls
	if startX > 1 {
		walls = append(walls, struct{ x, y int }{startX - 1, startY})
	}
	if startY > 1 {
		walls = append(walls, struct{ x, y int }{startX, startY - 1})
	}
	if startX < width-2 {
		walls = append(walls, struct{ x, y int }{startX + 1, startY})
	}
	if startY < height-2 {
		walls = append(walls, struct{ x, y int }{startX, startY + 1})
	}

	// Direction vectors for checking neighbors
	dx := []int{-1, 1, 0, 0}
	dy := []int{0, 0, -1, 1}

	// Process all walls
	for len(walls) > 0 {
		// Pick a random wall
		wallIndex := rand.Intn(len(walls))
		wallX, wallY := walls[wallIndex].x, walls[wallIndex].y

		// Remove the wall from the list
		walls[wallIndex] = walls[len(walls)-1]
		walls = walls[:len(walls)-1]

		// Count visited cells on either side of the wall
		visitedCount := 0
		emptyNeighborX, emptyNeighborY := -1, -1

		for i := 0; i < 4; i++ {
			nx, ny := wallX+dx[i], wallY+dy[i]
			if nx >= 0 && nx < width && ny >= 0 && ny < height && visited[ny][nx] {
				visitedCount++
				emptyNeighborX, emptyNeighborY = nx, ny
			}
		}

		// If exactly one side has been visited, make this wall a passage
		if visitedCount == 1 {
			// Make sure we're not removing outer walls
			if wallX > 0 && wallX < width-1 && wallY > 0 && wallY < height-1 {
				// Find the cell on the other side of the wall from the visited cell
				newCellX, newCellY := 2*wallX-emptyNeighborX, 2*wallY-emptyNeighborY

				// Check if the new cell is within bounds
				if newCellX >= 0 && newCellX < width && newCellY >= 0 && newCellY < height && !visited[newCellY][newCellX] {
					// Remove the wall (make it a passage)
					d.Cells[wallY][wallX] = Cell{Type: Empty}
					visited[wallY][wallX] = true

					// Mark the new cell as empty
					d.Cells[newCellY][newCellX] = Cell{Type: Empty}
					visited[newCellY][newCellX] = true

					// Add the walls of the new cell to the wall list
					for i := 0; i < 4; i++ {
						nx, ny := newCellX+dx[i], newCellY+dy[i]
						if nx >= 0 && nx < width && ny >= 0 && ny < height && !visited[ny][nx] && d.Cells[ny][nx].Type == Wall {
							walls = append(walls, struct{ x, y int }{nx, ny})
						}
					}
				}
			}
		}
	}

	// Create some additional random connections to make the maze less tree-like
	// This adds loops to the maze for more interesting gameplay
	numExtraConnections := (width * height) / 50
	for i := 0; i < numExtraConnections; i++ {
		x, y := rand.Intn(width-2)+1, rand.Intn(height-2)+1
		if d.Cells[y][x].Type == Wall {
			// Count neighboring passages
			passageCount := 0
			for j := 0; j < 4; j++ {
				nx, ny := x+dx[j], y+dy[j]
				if nx >= 0 && nx < width && ny >= 0 && ny < height && d.Cells[ny][nx].Type == Empty {
					passageCount++
				}
			}

			// If at least two neighbors are passages, remove this wall
			if passageCount >= 2 {
				d.Cells[y][x] = Cell{Type: Empty}
			}
		}
	}
}

func isWithinFOV(px, py, x, y, radius int) bool {
	dx := px - x
	dy := py - y
	return dx*dx+dy*dy <= radius*radius // Circular FOV
}

func (d *Dungeon) Draw(screen *ebiten.Image, player *Player) {
	for y, row := range d.Cells {
		for x, cell := range row {
			if player.FOVEnabled && !isWithinFOV(player.X, player.Y, x, y, player.FOVRadius) {
				continue // Skip drawing tiles outside the FOV
			}

			var clr color.RGBA
			switch cell.Type {
			case Empty:
				clr = color.RGBA{30, 30, 30, 255}
			case Wall:
				clr = color.RGBA{0, 0, 0, 255}
			case Monster:
				clr = color.RGBA{255, 0, 0, 255}
			case Treasure:
				clr = color.RGBA{255, 215, 0, 255}
			case Entrance:
				clr = color.RGBA{0, 255, 0, 255}
			case Exit:
				clr = color.RGBA{0, 0, 255, 255}
			}

			vector.DrawFilledRect(screen, float32(x*tileSize), float32(y*tileSize), float32(tileSize), float32(tileSize), clr, false)
		}
	}
}
