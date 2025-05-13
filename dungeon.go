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

func (ct CellType) String() string {
	switch ct {
	case Empty:
		return "Empty"
	case Wall:
		return "Wall"
	case Monster:
		return "Monster"
	case Treasure:
		return "Treasure"
	case Entrance:
		return "Entrance"
	case Exit:
		return "Exit"
	default:
		return "Unknown"
	}
}

// Optional: More structured data for monster & treasure classification
type TreasureType string

const (
	TreasureGold     TreasureType = "gold"
	TreasureGems     TreasureType = "gems"
	TreasureArtifact TreasureType = "artifact"
	TreasurePotion   TreasureType = "potion"
)

type MonsterTier int

const (
	TierEasy MonsterTier = iota
	TierMedium
	TierHard
	TierBoss
)

type Cell struct {
	Type             CellType
	InteractionLevel int          // Difficulty (monster) or value (treasure)
	TreasureType     TreasureType // Specific treasure variant
	MonsterTier      MonsterTier  // Optional: Add more scaling/behavior if needed
}

type Dungeon struct {
	Cells         [][]Cell
	Width, Height int
	Entrance      [2]int
	Exit          [2]int
	Visited       [][]bool
	Level         int
}

const (
	NumMonsters  = 10
	NumTreasures = 10
)

// Modify the NewDungeon function to initialize monsters and treasures with levels
func NewDungeon(width, height int, level int) *Dungeon {
	d := &Dungeon{
		Cells:   make([][]Cell, height),
		Width:   width,
		Height:  height,
		Visited: make([][]bool, height),
		Level:   level,
	}
	// initialize Cells and Visited
	for y := 0; y < height; y++ {
		d.Cells[y] = make([]Cell, width)
		d.Visited[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			d.Cells[y][x] = Cell{Type: Wall, InteractionLevel: 0, TreasureType: ""}
			d.Visited[y][x] = false
		}
	}

	// Generate maze with proper paths
	d.generateMaze()

	// Place entrance
	entranceX, entranceY := d.placeRandomFeature(Empty, Entrance)
	d.Entrance = [2]int{entranceX, entranceY}

	// Find dead ends that are far from the entrance
	deadEnds := d.findDeadEnds()

	// Sort dead ends by distance from entrance (in descending order)
	entrancePoint := [2]int{entranceX, entranceY}
	d.sortDeadEndsByDistance(deadEnds, entrancePoint)

	// Place exit at the furthest dead end
	if len(deadEnds) > 0 {
		exitX, exitY := deadEnds[0][0], deadEnds[0][1]
		d.Cells[exitY][exitX] = Cell{Type: Exit, InteractionLevel: level + 1}
		d.Exit = [2]int{exitX, exitY}
	} else {
		// Fallback if no suitable dead ends found
		var exitX, exitY int
		for {
			exitX, exitY = d.placeRandomFeature(Empty, Exit)
			// Check if exit is at least 1/3 of the dungeon size away from the entrance
			minDistance := (width + height) / 3
			dx, dy := entranceX-exitX, entranceY-exitY
			distance := dx*dx + dy*dy
			if distance >= minDistance*minDistance {
				d.Cells[exitY][exitX].InteractionLevel = level + 1
				d.Exit = [2]int{exitX, exitY}
				break
			}
			// Revert and try again
			d.Cells[exitY][exitX] = Cell{Type: Empty}
		}
	}

	// Place monsters with varying levels and tiers based on dungeon level
	for i := 0; i < NumMonsters; i++ {
		x, y := d.placeRandomFeature(Empty, Monster)

		// Monster level and tier logic
		monsterLevel := level + rand.Intn(3) - 1
		if monsterLevel < 1 {
			monsterLevel = 1
		}

		var tier MonsterTier
		switch {
		case monsterLevel <= 2:
			tier = TierEasy
		case monsterLevel <= 4:
			tier = TierMedium
		case monsterLevel <= 6:
			tier = TierHard
		default:
			tier = TierBoss
		}

		d.Cells[y][x].InteractionLevel = monsterLevel
		d.Cells[y][x].MonsterTier = tier
	}

	// Place treasures with type-safe treasure types
	treasureTypes := []TreasureType{TreasureGold, TreasureGems, TreasureArtifact, TreasurePotion}
	for i := 0; i < NumTreasures; i++ {
		x, y := d.placeRandomFeature(Empty, Treasure)

		treasureValue := level*10 + rand.Intn(20) - 10
		if treasureValue < 10 {
			treasureValue = 10
		}

		treasureType := treasureTypes[rand.Intn(len(treasureTypes))]

		d.Cells[y][x].InteractionLevel = treasureValue
		d.Cells[y][x].TreasureType = treasureType
	}

	return d
}

// Find all dead ends in the dungeon (empty cells with only one neighboring empty cell)
func (d *Dungeon) findDeadEnds() [][2]int {
	// Directions for checking neighbors (up, right, down, left)
	dirs := []struct{ dx, dy int }{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	deadEnds := [][2]int{}

	for y := 1; y < d.Height-1; y++ {
		for x := 1; x < d.Width-1; x++ {
			// Skip if not empty
			if d.Cells[y][x].Type != Empty {
				continue
			}

			// Count empty neighbors
			emptyNeighbors := 0
			for _, dir := range dirs {
				nx, ny := x+dir.dx, y+dir.dy
				if nx >= 0 && nx < d.Width && ny >= 0 && ny < d.Height && d.Cells[ny][nx].Type == Empty {
					emptyNeighbors++
				}
			}

			// If only one empty neighbor, this is a dead end
			if emptyNeighbors == 1 {
				deadEnds = append(deadEnds, [2]int{x, y})
			}
		}
	}

	return deadEnds
}

// Sort dead ends by distance from a point (descending order - farthest first)
func (d *Dungeon) sortDeadEndsByDistance(deadEnds [][2]int, point [2]int) {
	// Simple bubble sort based on distance
	for i := 0; i < len(deadEnds)-1; i++ {
		for j := 0; j < len(deadEnds)-i-1; j++ {
			dist1 := (deadEnds[j][0]-point[0])*(deadEnds[j][0]-point[0]) +
				(deadEnds[j][1]-point[1])*(deadEnds[j][1]-point[1])
			dist2 := (deadEnds[j+1][0]-point[0])*(deadEnds[j+1][0]-point[0]) +
				(deadEnds[j+1][1]-point[1])*(deadEnds[j+1][1]-point[1])

			// Sort in descending order (furthest first)
			if dist1 < dist2 {
				deadEnds[j], deadEnds[j+1] = deadEnds[j+1], deadEnds[j]
			}
		}
	}
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

type Point struct{ x, y int }

// Generates a randomized maze within the dungeon. (Randomized Prim’s Algorithm)
func (d *Dungeon) generateMaze() {
	// Initialize all cells as walls
	d.fillWithWalls()

	// Directions for movement: left, right, up, down (in steps of 2 for maze structure)
	dirs := []Point{{-2, 0}, {2, 0}, {0, -2}, {0, 2}}
	start := Point{1, 1}
	d.setCellEmpty(start)

	// Initialize the wall list from the start point's neighbors
	walls := d.getInitialWalls(start, dirs)

	// Main loop: continue until no walls are left to process
	for len(walls) > 0 {
		wall := d.randomWall(&walls) // Pick and remove a random wall

		// Skip if this wall is already part of the path
		if d.isEmpty(wall) {
			continue
		}

		// Get valid empty neighbors
		neighbors := d.getEmptyNeighbors(wall, dirs)
		if len(neighbors) > 0 {
			// Connect the wall with a randomly chosen neighbor
			neighbor := neighbors[rand.Intn(len(neighbors))]
			d.carvePath(wall, neighbor)

			// Add adjacent walls of the current wall to the list
			d.addAdjacentWalls(wall, dirs, &walls)
		}
	}
}

// Fill the entire dungeon with walls.
func (d *Dungeon) fillWithWalls() {
	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			d.Cells[y][x].Type = Wall
		}
	}
}

// Set a cell to be empty.
func (d *Dungeon) setCellEmpty(p Point) {
	d.Cells[p.y][p.x].Type = Empty
}

// Get the initial wall list from the start point's neighbors.
func (d *Dungeon) getInitialWalls(start Point, dirs []Point) []Point {
	walls := []Point{}
	for _, dir := range dirs {
		nextX, nextY := start.x+dir.x, start.y+dir.y
		if inBounds(nextX, nextY, d.Width, d.Height) {
			walls = append(walls, Point{nextX, nextY})
		}
	}
	return walls
}

// Randomly select and remove a wall from the list.
func (d *Dungeon) randomWall(walls *[]Point) Point {
	idx := rand.Intn(len(*walls))
	wall := (*walls)[idx]
	*walls = removeAt(*walls, idx) // Remove selected wall
	return wall
}

// Check if a cell is empty.
func (d *Dungeon) isEmpty(p Point) bool {
	return d.Cells[p.y][p.x].Type != Wall
}

// Get valid empty neighbors of a wall.
func (d *Dungeon) getEmptyNeighbors(wall Point, dirs []Point) []Point {
	var neighbors []Point
	for _, dir := range dirs {
		nextX, nextY := wall.x+dir.x, wall.y+dir.y
		if inBounds(nextX, nextY, d.Width, d.Height) && d.Cells[nextY][nextX].Type == Empty {
			neighbors = append(neighbors, Point{nextX, nextY})
		}
	}
	return neighbors
}

// Carve a path by removing a wall and the mid cell between wall and neighbor.
func (d *Dungeon) carvePath(wall, neighbor Point) {
	midX := (wall.x + neighbor.x) / 2
	midY := (wall.y + neighbor.y) / 2
	d.setCellEmpty(wall)
	d.setCellEmpty(Point{midX, midY})
}

// Add adjacent walls of a given wall to the list.
func (d *Dungeon) addAdjacentWalls(wall Point, dirs []Point, walls *[]Point) {
	for _, dir := range dirs {
		nextX, nextY := wall.x+dir.x, wall.y+dir.y
		if inBounds(nextX, nextY, d.Width, d.Height) && d.Cells[nextY][nextX].Type == Wall {
			*walls = append(*walls, Point{nextX, nextY})
		}
	}
}

func inBounds(x, y, width, height int) bool {
	return x >= 0 && x < width && y >= 0 && y < height
}

func removeAt(points []Point, i int) []Point {
	return append(points[:i], points[i+1:]...)
}

func isWithinFOV(px, py, x, y, radius int) bool {
	dx := px - x
	dy := py - y
	return dx*dx+dy*dy <= radius*radius // Circular FOV
}

func (d *Dungeon) Draw(screen *ebiten.Image, player *Player) {
	for y, row := range d.Cells {
		for x, cell := range row {
			withinFOV := isWithinFOV(player.X, player.Y, x, y, player.FOVRadius)

			// Skip drawing if not visible and never visited
			if player.FOVEnabled && !withinFOV && !d.Visited[y][x] {
				continue
			}

			// Mark as visited if within FOV
			if withinFOV {
				d.Visited[y][x] = true
			}

			clr := getCellColor(cell.Type, withinFOV)

			// Darken tile if seen before but not in current FOV
			if player.FOVEnabled && !withinFOV {
				clr = darkenColor(clr)
			}

			vector.DrawFilledRect(
				screen,
				float32(x*tileSize),
				float32(y*tileSize),
				float32(tileSize),
				float32(tileSize),
				clr,
				false,
			)
		}
	}
}
func getCellColor(cellType CellType, visible bool) color.RGBA {
	dimColor := color.RGBA{30, 30, 30, 255}

	if !visible {
		// Return dimmed default for hidden tiles
		switch cellType {
		case Monster, Treasure, Exit:
			return dimColor
		}
	}

	switch cellType {
	case Empty:
		return color.RGBA{30, 30, 30, 255}
	case Wall:
		return color.RGBA{0, 0, 0, 255}
	case Monster:
		return color.RGBA{255, 0, 0, 255}
	case Treasure:
		return color.RGBA{255, 215, 0, 255}
	case Entrance:
		return color.RGBA{0, 255, 0, 255}
	case Exit:
		return color.RGBA{0, 0, 255, 255}
	default:
		return color.RGBA{255, 255, 255, 255} // fallback
	}
}

func darkenColor(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: c.R / 2,
		G: c.G / 2,
		B: c.B / 2,
		A: c.A,
	}
}

func (d *Dungeon) FindPath(start, goal Point) []Point {
	type Node struct {
		Pos   Point
		Steps int
		Prev  *Node
	}

	width, height := d.Width, d.Height
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	queue := []*Node{{Pos: start}}
	var goalNode *Node

	dirs := []Point{
		{0, -1}, {1, 0}, {0, 1}, {-1, 0},
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		x, y := current.Pos.x, current.Pos.y
		if x == goal.x && y == goal.y {
			goalNode = current
			break
		}

		for _, dir := range dirs {
			nx, ny := x+dir.x, y+dir.y
			if nx >= 0 && ny >= 0 && nx < width && ny < height &&
				!visited[ny][nx] &&
				d.Cells[ny][nx].Type != Wall {

				visited[ny][nx] = true
				queue = append(queue, &Node{
					Pos:   Point{nx, ny},
					Prev:  current,
					Steps: current.Steps + 1,
				})
			}
		}
	}

	if goalNode == nil {
		return nil
	}

	// Reconstruct path
	var path []Point
	for node := goalNode; node != nil; node = node.Prev {
		path = append([]Point{node.Pos}, path...)
	}
	return path
}
