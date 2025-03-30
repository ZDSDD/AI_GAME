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
	Monster
	Treasure
	Entrance
	Exit
)

type Cell struct {
	Type CellType
}

type Dungeon struct {
	Cells [][]Cell
}

func NewDungeon(width, height int) *Dungeon {
	d := &Dungeon{Cells: make([][]Cell, height)}

	for y := 0; y < height; y++ {
		d.Cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			d.Cells[y][x] = Cell{Type: Empty}
		}
	}

	// Place entrance, exit, and some random treasures/monsters
	ex, ey := rand.Intn(width-2)+1, rand.Intn(height-2)+1
	d.Cells[ey][ex] = Cell{Type: Entrance}

	exitX, exitY := rand.Intn(width-2)+1, rand.Intn(height-2)+1
	d.Cells[exitY][exitX] = Cell{Type: Exit}

	for i := 0; i < 5; i++ {
		tx, ty := rand.Intn(width), rand.Intn(height)
		d.Cells[ty][tx] = Cell{Type: Treasure}

		mx, my := rand.Intn(width), rand.Intn(height)
		d.Cells[my][mx] = Cell{Type: Monster}
	}

	return d
}

func (d *Dungeon) FindEntrance() (int, int) {
	for y, row := range d.Cells {
		for x, cell := range row {
			if cell.Type == Entrance {
				return x, y
			}
		}
	}
	return 0, 0
}

func (d *Dungeon) Draw(screen *ebiten.Image) {
	for y, row := range d.Cells {
		for x, cell := range row {
			var clr color.RGBA

			switch cell.Type {
			case Empty:
				clr = color.RGBA{80, 80, 80, 255}
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
