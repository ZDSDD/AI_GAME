package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// GameState represents the current state of the game
type GameState int

const (
	StateMenu GameState = iota
	StateGame
)

// Define available resolution options
type Resolution struct {
	Width  int
	Height int
	Label  string
}

var resolutions = []Resolution{
	{800, 600, "800x600"},
	{1024, 768, "1024x768"},
	{1280, 720, "1280x720 (HD)"},
	{1366, 768, "1366x768"},
	{1600, 900, "1600x900"},
	{1920, 1080, "1920x1080 (Full HD)"},
}

// Define default tile sizes options
var tileSizeOptions = []int{8, 12, 16, 20, 24, 32}

// Define difficulty options
type Difficulty struct {
	Level       int
	Label       string
	MonsterMod  float64 // Monster strength modifier
	TreasureMod float64 // Treasure value modifier
}

var difficulties = []Difficulty{
	{1, "Easy", 0.8, 1.2},
	{2, "Normal", 1.0, 1.0},
	{3, "Hard", 1.2, 0.8},
	{4, "Nightmare", 1.5, 0.7},
}

// Button represents a clickable UI element
type Button struct {
	X, Y          int
	Width, Height int
	Label         string
	Selected      bool
	OnClick       func()
}

// MainMenu represents the pre-game options panel
type MainMenu struct {
	selectedResolution int
	selectedTileSize   int
	selectedDifficulty int
	enableFOV          bool
	dungeonWidth       int
	dungeonHeight      int
	buttons            []*Button
	sliders            []*Slider

	// Scroll related properties
	scrollY       int  // Current scroll position
	contentHeight int  // Total height of all content
	scrollBarGrab bool // Is the scrollbar being dragged
	scrollBarY    int  // Y position of the scrollbar
}

// Slider represents a UI element for selecting a numeric value
type Slider struct {
	X, Y          int
	Width, Height int
	Label         string
	MinValue      int
	MaxValue      int
	Value         int
	OnChange      func(int)
	Active        bool // Is the slider actively being dragged
}

// GameSettings contains all settings for the game
type GameSettings struct {
	ScreenWidth    int
	ScreenHeight   int
	TileSize       int
	DungeonWidth   int
	DungeonHeight  int
	EnableFOV      bool
	DifficultyMods struct {
		Monster  float64
		Treasure float64
	}
}

// MainGame is the root game struct that manages game state
type MainGame struct {
	state    GameState
	menu     *MainMenu
	game     *Game
	settings GameSettings
}

func NewMainGame() *MainGame {
	menu := &MainMenu{
		selectedResolution: 2, // Default to 1280x720
		selectedTileSize:   2, // Default to 16
		selectedDifficulty: 1, // Default to Normal
		enableFOV:          true,
		dungeonWidth:       40, // Default width
		dungeonHeight:      20, // Default height
		scrollY:            0,
	}

	// Default settings
	settings := GameSettings{
		ScreenWidth:   resolutions[menu.selectedResolution].Width,
		ScreenHeight:  resolutions[menu.selectedResolution].Height,
		TileSize:      tileSizeOptions[menu.selectedTileSize],
		DungeonWidth:  menu.dungeonWidth,
		DungeonHeight: menu.dungeonHeight,
		EnableFOV:     menu.enableFOV,
	}
	settings.DifficultyMods.Monster = difficulties[menu.selectedDifficulty].MonsterMod
	settings.DifficultyMods.Treasure = difficulties[menu.selectedDifficulty].TreasureMod

	mainGame := &MainGame{
		state:    StateMenu,
		menu:     menu,
		settings: settings,
	}

	// Initialize menu buttons
	mainGame.initializeMenu()

	return mainGame
}

// Initialize menu elements
func (m *MainGame) initializeMenu() {
	buttonY := 140
	buttonSpacing := 40

	m.menu.buttons = []*Button{}

	// Title section doesn't need to be a button, it will be drawn separately

	// Resolution section
	buttonY += buttonSpacing
	resolutionLabel := &Button{
		X:        m.settings.ScreenWidth/2 - 150,
		Y:        buttonY,
		Width:    300,
		Height:   30,
		Label:    "Display Resolution",
		Selected: false,
	}
	m.menu.buttons = append(m.menu.buttons, resolutionLabel)

	buttonY += 35
	resolutionButtons := []*Button{}
	for i, res := range resolutions {
		resIndex := i // Capture the index for closure
		button := &Button{
			X:        m.settings.ScreenWidth/2 - 150,
			Y:        buttonY + i*35,
			Width:    300,
			Height:   30,
			Label:    res.Label,
			Selected: i == m.menu.selectedResolution,
			OnClick: func() {
				m.menu.selectedResolution = resIndex
				// Update all button selected states
				for j, btn := range m.menu.buttons {
					if strings.Contains(btn.Label, "x") { // Simple check for resolution buttons
						m.menu.buttons[j].Selected = (j-2 == resIndex) // Adjust index offset based on your buttons array
					}
				}
				m.updateSettings()
				m.initializeMenu() // Reinitialize the menu after changing resolution
			},
		}
		resolutionButtons = append(resolutionButtons, button)
	}
	m.menu.buttons = append(m.menu.buttons, resolutionButtons...)

	buttonY += len(resolutions)*35 + buttonSpacing

	// Tile size buttons
	tileSizeLabel := &Button{
		X:        m.settings.ScreenWidth/2 - 150,
		Y:        buttonY,
		Width:    300,
		Height:   30,
		Label:    "Tile Size",
		Selected: false,
	}
	m.menu.buttons = append(m.menu.buttons, tileSizeLabel)

	buttonY += 35
	tileSizeButtons := []*Button{}
	for i, size := range tileSizeOptions {
		sizeIndex := i // Capture the index for closure
		button := &Button{
			X:        m.settings.ScreenWidth/2 - 150 + (i%3)*100,
			Y:        buttonY + (i/3)*35,
			Width:    90,
			Height:   30,
			Label:    fmt.Sprintf("%dpx", size),
			Selected: i == m.menu.selectedTileSize,
			OnClick: func() {
				// Deselect all tile size buttons first
				for j, btn := range m.menu.buttons {
					if strings.HasSuffix(btn.Label, "px") {
						m.menu.buttons[j].Selected = false
					}
				}

				// Now select this button
				for j, btn := range m.menu.buttons {
					if btn.Label == fmt.Sprintf("%dpx", tileSizeOptions[sizeIndex]) {
						m.menu.buttons[j].Selected = true
						break
					}
				}

				m.menu.selectedTileSize = sizeIndex
				m.updateSettings()
			},
		}
		tileSizeButtons = append(tileSizeButtons, button)
	}
	m.menu.buttons = append(m.menu.buttons, tileSizeButtons...)

	buttonY += 70 + buttonSpacing

	// Difficulty buttons
	difficultyLabel := &Button{
		X:        m.settings.ScreenWidth/2 - 150,
		Y:        buttonY,
		Width:    300,
		Height:   30,
		Label:    "Difficulty",
		Selected: false,
	}
	m.menu.buttons = append(m.menu.buttons, difficultyLabel)

	buttonY += 35
	difficultyButtons := []*Button{}
	for i, diff := range difficulties {
		diffIndex := i // Capture the index for closure
		button := &Button{
			X:        m.settings.ScreenWidth/2 - 150 + (i%2)*150,
			Y:        buttonY + (i/2)*35,
			Width:    140,
			Height:   30,
			Label:    diff.Label,
			Selected: i == m.menu.selectedDifficulty,
			OnClick: func() {
				// Deselect all difficulty buttons first
				for j, btn := range m.menu.buttons {
					for _, d := range difficulties {
						if btn.Label == d.Label {
							m.menu.buttons[j].Selected = false
						}
					}
				}

				// Now select this button
				for j, btn := range m.menu.buttons {
					if btn.Label == difficulties[diffIndex].Label {
						m.menu.buttons[j].Selected = true
						break
					}
				}

				m.menu.selectedDifficulty = diffIndex
				m.updateSettings()
			},
		}
		difficultyButtons = append(difficultyButtons, button)
	}
	m.menu.buttons = append(m.menu.buttons, difficultyButtons...)

	buttonY += 70 + buttonSpacing

	// FOV toggle button
	fovButton := &Button{
		X:      m.settings.ScreenWidth/2 - 150,
		Y:      buttonY,
		Width:  300,
		Height: 30,
		Label: fmt.Sprintf("Field of View: %v", func() string {
			if m.menu.enableFOV {
				return "ON"
			} else {
				return "OFF"
			}
		}()),
		Selected: m.menu.enableFOV,
		OnClick: func() {
			m.menu.enableFOV = !m.menu.enableFOV

			// Update this button's state and label
			for j, btn := range m.menu.buttons {
				if strings.HasPrefix(btn.Label, "Field of View:") {
					m.menu.buttons[j].Selected = m.menu.enableFOV
					m.menu.buttons[j].Label = fmt.Sprintf("Field of View: %v", func() string {
						if m.menu.enableFOV {
							return "ON"
						} else {
							return "OFF"
						}
					}())
					break
				}
			}

			m.updateSettings()
		},
	}
	m.menu.buttons = append(m.menu.buttons, fovButton)

	buttonY += buttonSpacing + 20

	// Dungeon size sliders
	dungeonWidthSlider := &Slider{
		X:        m.settings.ScreenWidth/2 - 150,
		Y:        buttonY,
		Width:    300,
		Height:   20,
		Label:    fmt.Sprintf("Dungeon Width: %d", m.menu.dungeonWidth),
		MinValue: 20,
		MaxValue: 80,
		Value:    m.menu.dungeonWidth,
		OnChange: func(val int) {
			m.menu.dungeonWidth = val
			m.updateSettings()
		},
	}

	buttonY += 50

	dungeonHeightSlider := &Slider{
		X:        m.settings.ScreenWidth/2 - 150,
		Y:        buttonY,
		Width:    300,
		Height:   20,
		Label:    fmt.Sprintf("Dungeon Height: %d", m.menu.dungeonHeight),
		MinValue: 10,
		MaxValue: 40,
		Value:    m.menu.dungeonHeight,
		OnChange: func(val int) {
			m.menu.dungeonHeight = val
			m.updateSettings()
		},
	}

	m.menu.sliders = []*Slider{dungeonWidthSlider, dungeonHeightSlider}

	buttonY += 70

	// Start Game button
	startButton := &Button{
		X:        m.settings.ScreenWidth/2 - 100,
		Y:        buttonY,
		Width:    200,
		Height:   40,
		Label:    "Start Game",
		Selected: false,
		OnClick: func() {
			m.startGame()
		},
	}
	m.menu.buttons = append(m.menu.buttons, startButton)

	// Calculate total content height for scrollbar
	m.menu.contentHeight = buttonY + 60 // Add some padding at the bottom
}

// Update the game settings based on menu selections
func (m *MainGame) updateSettings() {
	m.settings.ScreenWidth = resolutions[m.menu.selectedResolution].Width
	m.settings.ScreenHeight = resolutions[m.menu.selectedResolution].Height
	m.settings.TileSize = tileSizeOptions[m.menu.selectedTileSize]
	m.settings.DungeonWidth = m.menu.dungeonWidth
	m.settings.DungeonHeight = m.menu.dungeonHeight
	m.settings.EnableFOV = m.menu.enableFOV
	m.settings.DifficultyMods.Monster = difficulties[m.menu.selectedDifficulty].MonsterMod
	m.settings.DifficultyMods.Treasure = difficulties[m.menu.selectedDifficulty].TreasureMod

	// Update window size
	ebiten.SetWindowSize(m.settings.ScreenWidth, m.settings.ScreenHeight)
}

// Start the game with current settings
func (m *MainGame) startGame() {
	// Create a new game with the selected settings
	dungeon := NewDungeon(m.settings.DungeonWidth, m.settings.DungeonHeight, difficulties[m.menu.selectedDifficulty].Level)
	player := NewPlayer(dungeon.Entrance)
	player.FOVEnabled = m.settings.EnableFOV

	// Create the interaction handler with difficulty modifiers
	interactionHandler := NewInteractionHandler()

	// Register interactions for different cell types with difficulty modifiers
	interactionHandler.Register(Monster, NewMonsterInteraction(1))            // Will be overridden per cell
	interactionHandler.Register(Treasure, NewTreasureInteraction(10, "gold")) // Will be overridden per cell
	interactionHandler.Register(Exit, NewExitInteraction(2))                  // Go to level 2

	m.game = &Game{
		dungeon:            dungeon,
		player:             player,
		interactionHandler: interactionHandler,
	}

	// Apply difficulty modifiers to monsters and treasures
	for y := 0; y < dungeon.Height; y++ {
		for x := 0; x < dungeon.Width; x++ {
			cell := &dungeon.Cells[y][x]
			if cell.Type == Monster {
				cell.InteractionLevel = int(float64(cell.InteractionLevel) * m.settings.DifficultyMods.Monster)
				if cell.InteractionLevel < 1 {
					cell.InteractionLevel = 1
				}
			} else if cell.Type == Treasure {
				cell.InteractionLevel = int(float64(cell.InteractionLevel) * m.settings.DifficultyMods.Treasure)
				if cell.InteractionLevel < 5 {
					cell.InteractionLevel = 5 // Minimum treasure value
				}
			}
		}
	}

	m.state = StateGame

	// Set global tileSize variable used in other files
	tileSize = m.settings.TileSize
}

// Use the standard library strings package for string operations

func (m *MainGame) Update() error {
	switch m.state {
	case StateMenu:
		mouseX, mouseY := ebiten.CursorPosition()

		// Handle scrolling with mouse wheel
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 {
			m.menu.scrollY -= int(wheelY * 20)
			// Clamp scrolling
			maxScroll := m.menu.contentHeight - m.settings.ScreenHeight + 40
			if maxScroll < 0 {
				maxScroll = 0
			}

			if m.menu.scrollY < 0 {
				m.menu.scrollY = 0
			} else if m.menu.scrollY > maxScroll {
				m.menu.scrollY = maxScroll
			}
		}

		// Calculate scrollbar properties
		viewportHeight := m.settings.ScreenHeight
		scrollBarHeight := int(float64(viewportHeight) * float64(viewportHeight) / float64(m.menu.contentHeight))
		if scrollBarHeight < 30 {
			scrollBarHeight = 30 // Minimum height for visibility
		}

		// Calculate scrollbar position
		maxScroll := m.menu.contentHeight - viewportHeight + 40
		if maxScroll <= 0 {
			m.menu.scrollBarY = 0
		} else {
			m.menu.scrollBarY = int(float64(m.menu.scrollY) / float64(maxScroll) * float64(viewportHeight-scrollBarHeight))
		}

		// Handle scrollbar dragging
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// Check if clicked on scrollbar
			scrollBarX := m.settings.ScreenWidth - 20
			scrollBarWidth := 10

			if !m.menu.scrollBarGrab && mouseX >= scrollBarX && mouseX <= scrollBarX+scrollBarWidth &&
				mouseY >= m.menu.scrollBarY && mouseY <= m.menu.scrollBarY+scrollBarHeight {
				m.menu.scrollBarGrab = true
			}

			if m.menu.scrollBarGrab {
				// Calculate scroll position based on mouse position
				maxScrollBarPos := viewportHeight - scrollBarHeight
				scrollPct := float64(mouseY) / float64(maxScrollBarPos)
				if scrollPct < 0 {
					scrollPct = 0
				}
				if scrollPct > 1 {
					scrollPct = 1
				}

				maxScroll := m.menu.contentHeight - viewportHeight + 40
				m.menu.scrollY = int(scrollPct * float64(maxScroll))
				if m.menu.scrollY < 0 {
					m.menu.scrollY = 0
				} else if m.menu.scrollY > maxScroll {
					m.menu.scrollY = maxScroll
				}
			}
		} else {
			m.menu.scrollBarGrab = false
		}

		// Handle mouse button clicks on UI elements
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			// Adjust mouse Y position for scrolling
			adjustedMouseY := mouseY + m.menu.scrollY

			// Check button clicks
			for _, button := range m.menu.buttons {
				if button.OnClick != nil &&
					mouseX >= button.X && mouseX < button.X+button.Width &&
					adjustedMouseY >= button.Y && adjustedMouseY < button.Y+button.Height {
					button.OnClick()
				}
			}

			// Check slider clicks
			for i, slider := range m.menu.sliders {
				if adjustedMouseY >= slider.Y && adjustedMouseY < slider.Y+slider.Height &&
					mouseX >= slider.X && mouseX < slider.X+slider.Width {
					// Calculate position within slider
					pos := float64(mouseX-slider.X) / float64(slider.Width)
					newVal := slider.MinValue + int(pos*float64(slider.MaxValue-slider.MinValue))
					if newVal < slider.MinValue {
						newVal = slider.MinValue
					}
					if newVal > slider.MaxValue {
						newVal = slider.MaxValue
					}
					slider.Value = newVal
					slider.Active = true

					// Update slider label
					if i == 0 {
						slider.Label = fmt.Sprintf("Dungeon Width: %d", newVal)
						m.menu.dungeonWidth = newVal
					} else if i == 1 {
						slider.Label = fmt.Sprintf("Dungeon Height: %d", newVal)
						m.menu.dungeonHeight = newVal
					}

					if slider.OnChange != nil {
						slider.OnChange(newVal)
					}
				}
			}
		}

		// Update active sliders even when mouse is held
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for i, slider := range m.menu.sliders {
				if slider.Active {
					// Calculate position within slider
					pos := float64(mouseX-slider.X) / float64(slider.Width)
					newVal := slider.MinValue + int(pos*float64(slider.MaxValue-slider.MinValue))
					newVal = int(math.Max(float64(slider.MinValue), math.Min(float64(slider.MaxValue), float64(newVal))))
					slider.Value = newVal

					// Update slider label and value
					if i == 0 {
						slider.Label = fmt.Sprintf("Dungeon Width: %d", newVal)
						m.menu.dungeonWidth = newVal
					} else if i == 1 {
						slider.Label = fmt.Sprintf("Dungeon Height: %d", newVal)
						m.menu.dungeonHeight = newVal
					}

					if slider.OnChange != nil {
						slider.OnChange(newVal)
					}
				}
			}
		} else {
			// Reset active state when mouse button is released
			for i := range m.menu.sliders {
				m.menu.sliders[i].Active = false
			}
		}

	case StateGame:
		if m.game != nil {
			return m.game.Update()
		}
	}

	return nil
}

func (m *MainGame) Draw(screen *ebiten.Image) {
	switch m.state {
	case StateMenu:
		// Draw background
		screen.Fill(color.RGBA{20, 20, 30, 255})

		// Create a clipping area for scrolling content
		clipY := 0
		clipHeight := m.settings.ScreenHeight

		// Draw title (always visible, doesn't scroll)
		titleText := "Procedural Dungeon - Game Options"
		titleX := m.settings.ScreenWidth/2 - len(titleText)*4
		ebitenutil.DebugPrintAt(screen, titleText, titleX, 80)

		// Draw scrollable content
		for _, button := range m.menu.buttons {
			// Adjust y position for scrolling
			adjY := button.Y - m.menu.scrollY

			// Skip rendering if outside the viewport
			if adjY+button.Height < clipY || adjY > clipY+clipHeight {
				continue
			}

			// Skip buttons that are just labels
			if button.OnClick == nil {
				ebitenutil.DebugPrintAt(screen, button.Label, button.X+10, adjY+10)
				continue
			}

			// Draw button background
			bgColor := color.RGBA{50, 50, 60, 255}
			if button.Selected {
				bgColor = color.RGBA{100, 100, 200, 255}
			}

			vector.DrawFilledRect(screen, float32(button.X), float32(adjY),
				float32(button.Width), float32(button.Height), bgColor, false)

			// Draw button border
			borderColor := color.RGBA{200, 200, 220, 255}
			vector.StrokeRect(screen, float32(button.X), float32(adjY),
				float32(button.Width), float32(button.Height), 1, borderColor, false)

			// Draw button text
			ebitenutil.DebugPrintAt(screen, button.Label, button.X+10, adjY+10)
		}

		// Draw sliders
		for _, slider := range m.menu.sliders {
			// Adjust y position for scrolling
			adjY := slider.Y - m.menu.scrollY

			// Skip rendering if outside the viewport
			if adjY+slider.Height < clipY || adjY > clipY+clipHeight {
				continue
			}

			// Draw slider label
			ebitenutil.DebugPrintAt(screen, slider.Label, slider.X, adjY-15)

			// Draw slider track
			trackColor := color.RGBA{80, 80, 90, 255}
			vector.DrawFilledRect(screen, float32(slider.X), float32(adjY),
				float32(slider.Width), float32(slider.Height), trackColor, false)

			// Draw slider handle
			handlePos := float32(slider.X) + float32(slider.Width)*
				float32(slider.Value-slider.MinValue)/float32(slider.MaxValue-slider.MinValue)
			handleColor := color.RGBA{180, 180, 220, 255}
			vector.DrawFilledRect(screen,
				handlePos-5, float32(adjY)-5,
				10, float32(slider.Height)+10,
				handleColor, false)
		}

		// Draw scrollbar if content is larger than viewport
		if m.menu.contentHeight > m.settings.ScreenHeight {
			scrollBarX := m.settings.ScreenWidth - 20
			scrollBarWidth := 10

			// Calculate scrollbar height and position
			viewportHeight := m.settings.ScreenHeight
			scrollBarHeight := int(float64(viewportHeight) * float64(viewportHeight) / float64(m.menu.contentHeight))
			if scrollBarHeight < 30 {
				scrollBarHeight = 30 // Minimum height for visibility
			}

			// Draw scrollbar track
			trackColor := color.RGBA{40, 40, 50, 255}
			vector.DrawFilledRect(screen, float32(scrollBarX), 0,
				float32(scrollBarWidth), float32(viewportHeight), trackColor, false)

			// Draw scrollbar handle
			handleColor := color.RGBA{100, 100, 120, 255}
			if m.menu.scrollBarGrab {
				handleColor = color.RGBA{120, 120, 150, 255}
			}
			vector.DrawFilledRect(screen, float32(scrollBarX), float32(m.menu.scrollBarY),
				float32(scrollBarWidth), float32(scrollBarHeight), handleColor, false)
		}

	case StateGame:
		if m.game != nil {
			m.game.Draw(screen)
		}
	}
}

func (m *MainGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return m.settings.ScreenWidth, m.settings.ScreenHeight
}

func Contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func HasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	return s[:len(prefix)] == prefix
}

func HasSuffix(s, suffix string) bool {
	if len(suffix) > len(s) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
