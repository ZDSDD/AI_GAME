package main

import (
	"fmt"
	"time"
)

// --- Message with Timestamp ---

type TimedMessage struct {
	Text          string
	CreatedAt     time.Time
	TotalLifetime float64 // Message lifetime in seconds
	RemainingTime float64 // Remaining time before message disappears
}

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
	Messages     []TimedMessage
	MessageLife  float64 // Default lifetime for messages in seconds
}

func NewInteractionHandler() *InteractionHandler {
	return &InteractionHandler{
		Interactions: make(map[CellType]Interactable),
		Messages:     make([]TimedMessage, 0, 5),
		MessageLife:  1.5, // Default 1 second lifetime
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
	timedMsg := TimedMessage{
		Text:          msg,
		CreatedAt:     time.Now(),
		TotalLifetime: h.MessageLife,
		RemainingTime: h.MessageLife,
	}

	h.Messages = append(h.Messages, timedMsg)

	// Still cap the total number of messages to avoid memory buildup
	if len(h.Messages) > 5 {
		h.Messages = h.Messages[len(h.Messages)-5:]
	}
}

// UpdateMessages updates the remaining time for all messages and removes expired ones
func (h *InteractionHandler) UpdateMessages() {
	now := time.Now()
	var activeMessages []TimedMessage

	for _, msg := range h.Messages {
		elapsed := now.Sub(msg.CreatedAt).Seconds()
		remaining := h.MessageLife - elapsed

		if remaining > 0 {
			// Update the remaining time and keep the message
			msg.RemainingTime = remaining
			activeMessages = append(activeMessages, msg)
		}
		// If remaining <= 0, the message is not added to activeMessages
	}

	h.Messages = activeMessages
}

// GetActiveMessages returns only messages that haven't expired
func (h *InteractionHandler) GetActiveMessages() []TimedMessage {
	// Update the remaining time for all messages before returning
	h.UpdateMessages()
	return h.Messages
}

// For backward compatibility
func (h *InteractionHandler) GetMessages() []string {
	h.UpdateMessages()
	messages := make([]string, 0, len(h.Messages))
	for _, msg := range h.Messages {
		messages = append(messages, msg.Text)
	}
	return messages
}
