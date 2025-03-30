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
	NumMonsters  = 10
	NumTreasures = 10
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
	d.generateMaze(d.Width, d.Height)

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

func (d *Dungeon) generateMaze(width, height int) {
	// Initialize all cells as walls
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			d.Cells[y][x].Type = Wall
		}
	}

	// Directions for moving two cells (up, down, left, right)
	dirs := []struct{ dx, dy int }{{-2, 0}, {2, 0}, {0, -2}, {0, 2}}

	// Choose a starting cell (odd coordinates to align with rooms)
	startX, startY := 1, 1
	if startX >= width || startY >= height {
		startX, startY = 1, 1
	}
	d.Cells[startY][startX].Type = Empty

	// Initialize the list of frontier walls
	walls := []struct{ x, y int }{}
	for _, dir := range dirs {
		nx, ny := startX+dir.dx, startY+dir.dy
		if nx >= 0 && nx < width && ny >= 0 && ny < height {
			walls = append(walls, struct{ x, y int }{nx, ny})
		}
	}

	for len(walls) > 0 {
		// Pick a random wall from the frontier
		idx := rand.Intn(len(walls))
		wall := walls[idx]
		walls = append(walls[:idx], walls[idx+1:]...)

		// Check if the wall is still a wall
		if d.Cells[wall.y][wall.x].Type != Wall {
			continue
		}

		// Find visited neighbors (two cells away)
		var visitedNeighbors [][2]int
		for _, dir := range dirs {
			nx, ny := wall.x+dir.dx, wall.y+dir.dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height && d.Cells[ny][nx].Type == Empty {
				visitedNeighbors = append(visitedNeighbors, [2]int{nx, ny})
			}
		}

		if len(visitedNeighbors) > 0 {
			// Pick a random visited neighbor
			neighbor := visitedNeighbors[rand.Intn(len(visitedNeighbors))]
			// Carve the path between the wall and the neighbor
			midX := (wall.x + neighbor[0]) / 2
			midY := (wall.y + neighbor[1]) / 2
			d.Cells[wall.y][wall.x].Type = Empty
			d.Cells[midY][midX].Type = Empty

			// Add new frontier walls
			for _, dir := range dirs {
				nx, ny := wall.x+dir.dx, wall.y+dir.dy
				if nx >= 0 && nx < width && ny >= 0 && ny < height && d.Cells[ny][nx].Type == Wall {
					walls = append(walls, struct{ x, y int }{nx, ny})
				}
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
