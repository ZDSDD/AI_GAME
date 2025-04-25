package main

import "fmt"

// InteractionResult represents what happened after an interaction
type InteractionResult struct {
	Message       string
	HealthChange  int
	ScoreChange   int
	RemoveEntity  bool     // Whether the entity should be removed from the map
	EntityRemoved CellType // The type of entity that was removed (if any)
}

// Interactable is an interface for anything that can be interacted with
type Interactable interface {
	Interact(player *Player) InteractionResult
}

// MonsterInteraction handles interactions with monsters
type MonsterInteraction struct {
	Level int // Monster difficulty level
}

func NewMonsterInteraction(level int) *MonsterInteraction {
	return &MonsterInteraction{Level: level}
}

func (m *MonsterInteraction) Interact(player *Player) InteractionResult {
	// Simple combat: Player always wins but takes damage based on monster level
	damageToPlayer := 5 + (m.Level * 2)
	scoreGain := 10 + (m.Level * 5)

	// Apply the player's stats to modify combat results
	// This is where you could add more complex combat logic
	damageToPlayer = damageToPlayer * (100 - player.Defense) / 100

	return InteractionResult{
		Message:       fmt.Sprintf("Defeated a level %d monster! Took %d damage.", m.Level, damageToPlayer),
		HealthChange:  -damageToPlayer,
		ScoreChange:   scoreGain,
		RemoveEntity:  true,
		EntityRemoved: Monster,
	}
}

// TreasureInteraction handles interactions with treasures
type TreasureInteraction struct {
	Value int    // Value of the treasure
	Type  string // Type of treasure (gold, gem, artifact, etc.)
}

func NewTreasureInteraction(value int, treasureType string) *TreasureInteraction {
	return &TreasureInteraction{Value: value, Type: treasureType}
}

func (t *TreasureInteraction) Interact(player *Player) InteractionResult {
	scoreGain := t.Value

	// Apply player's luck to modify treasure value
	scoreGain = scoreGain * (100 + player.Luck) / 100

	// Bonus health for finding treasure
	healthGain := 0
	if t.Type == "potion" {
		healthGain = 10
	}

	return InteractionResult{
		Message:       fmt.Sprintf("Found %s worth %d points!", t.Type, scoreGain),
		HealthChange:  healthGain,
		ScoreChange:   scoreGain,
		RemoveEntity:  true,
		EntityRemoved: Treasure,
	}
}

// ExitInteraction handles interactions with the dungeon exit
type ExitInteraction struct {
	NextLevel int
}

func NewExitInteraction(nextLevel int) *ExitInteraction {
	return &ExitInteraction{NextLevel: nextLevel}
}

func (e *ExitInteraction) Interact(player *Player) InteractionResult {
	return InteractionResult{
		Message:       fmt.Sprintf("Descending to dungeon level %d!", e.NextLevel),
		HealthChange:  0,
		ScoreChange:   20, // Bonus for finding the exit
		RemoveEntity:  false,
		EntityRemoved: Empty, // Not removing anything
	}
}

// InteractionHandler manages all interactions in the game
type InteractionHandler struct {
	Interactions map[CellType]Interactable
	Messages     []string // Last few interaction messages
}

func NewInteractionHandler() *InteractionHandler {
	return &InteractionHandler{
		Interactions: make(map[CellType]Interactable),
		Messages:     []string{},
	}
}

// Register adds an interaction type to the handler
func (h *InteractionHandler) Register(cellType CellType, interaction Interactable) {
	h.Interactions[cellType] = interaction
}

// Handle processes an interaction with a specific cell
func (h *InteractionHandler) Handle(cellType CellType, player *Player) InteractionResult {
	// Check if we have a registered interaction for this cell type
	if interaction, exists := h.Interactions[cellType]; exists {
		result := interaction.Interact(player)

		// Store the message
		h.AddMessage(result.Message)

		// Apply changes to player
		player.Health += result.HealthChange
		player.Score += result.ScoreChange

		// Cap health at maximum
		if player.Health > player.MaxHealth {
			player.Health = player.MaxHealth
		}

		return result
	}

	// No interaction available
	return InteractionResult{
		Message:      "Nothing happens.",
		HealthChange: 0,
		ScoreChange:  0,
		RemoveEntity: false,
	}
}

// AddMessage adds a new message to the history
func (h *InteractionHandler) AddMessage(message string) {
	h.Messages = append(h.Messages, message)

	// Keep only the last 5 messages
	if len(h.Messages) > 5 {
		h.Messages = h.Messages[len(h.Messages)-5:]
	}
}

// GetMessages returns the most recent messages
func (h *InteractionHandler) GetMessages() []string {
	return h.Messages
}
