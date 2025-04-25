package main

import "fmt"

// --- Interaction Result ---

type InteractionResult struct {
	Message       string
	HealthChange  int
	ScoreChange   int
	RemoveEntity  bool
	EntityRemoved CellType
}

// --- Interactable Interface ---

type Interactable interface {
	Interact(player *Player) InteractionResult
}

// --- Monster Interaction ---

type MonsterInteraction struct {
	Level int
}

func NewMonsterInteraction(level int) *MonsterInteraction {
	return &MonsterInteraction{Level: level}
}

func (m *MonsterInteraction) Interact(player *Player) InteractionResult {
	damage := (5 + m.Level*2) * (100 - player.Defense) / 100
	score := 10 + m.Level*5

	return InteractionResult{
		Message:       fmt.Sprintf("Defeated a level %d monster! Took %d damage.", m.Level, damage),
		HealthChange:  -damage,
		ScoreChange:   score,
		RemoveEntity:  true,
		EntityRemoved: Monster,
	}
}

// --- Treasure Interaction ---

type TreasureInteraction struct {
	Value int
	Type  TreasureType
}

func NewTreasureInteraction(value int, ttype TreasureType) *TreasureInteraction {
	return &TreasureInteraction{Value: value, Type: ttype}
}

func (t *TreasureInteraction) Interact(player *Player) InteractionResult {
	score := t.Value * (100 + player.Luck) / 100
	health := 0

	if t.Type == TreasurePotion {
		health = 10
	}

	return InteractionResult{
		Message:       fmt.Sprintf("Found %s worth %d points!", t.Type, score),
		HealthChange:  health,
		ScoreChange:   score,
		RemoveEntity:  true,
		EntityRemoved: Treasure,
	}
}

// --- Exit Interaction ---

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
		ScoreChange:   20,
		RemoveEntity:  false,
		EntityRemoved: Empty,
	}
}

// --- Interaction Handler ---

type InteractionHandler struct {
	Interactions map[CellType]Interactable
	Messages     []string
}

func NewInteractionHandler() *InteractionHandler {
	return &InteractionHandler{
		Interactions: make(map[CellType]Interactable),
		Messages:     make([]string, 0, 5),
	}
}

func (h *InteractionHandler) Register(cellType CellType, i Interactable) {
	h.Interactions[cellType] = i
}

func (h *InteractionHandler) Handle(cellType CellType, player *Player) InteractionResult {
	if interaction, ok := h.Interactions[cellType]; ok {
		result := interaction.Interact(player)

		h.AddMessage(result.Message)
		player.Health += result.HealthChange
		player.Score += result.ScoreChange

		if player.Health > player.MaxHealth {
			player.Health = player.MaxHealth
		}

		return result
	}

	return InteractionResult{
		Message: "Nothing happens.",
	}
}

func (h *InteractionHandler) AddMessage(msg string) {
	h.Messages = append(h.Messages, msg)
	if len(h.Messages) > 5 {
		h.Messages = h.Messages[len(h.Messages)-5:]
	}
}

func (h *InteractionHandler) GetMessages() []string {
	return h.Messages
}
