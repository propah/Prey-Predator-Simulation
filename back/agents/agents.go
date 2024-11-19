package agents

import (
	"Prey_Predator_MAS/Brain"
	"Prey_Predator_MAS/config"
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/quartercastle/vector"
)

type Agent struct {
	ID          uint32        `json:"id"`
	Position    vector.Vector `json:"pos"`
	Color       string        `json:"color"`
	Velocity    vector.Vector `json:"-"`
	Perceipt    Perceipt      `json:"-"`
	RaysValues  []float64     `json:"-"`
	changedCell bool          `json:"-"`
	Brain       *Brain.Brain  `json:"-"`
	Speed       float64
	Rotation    float64

	LifePoints   int
	Energy       int
	Reproduction int
	Digestion    int

	Regen bool

	Generation int

	lock sync.Mutex
}

type AgentViewModel struct {
	ID         uint32                `json:"id"`
	Position   vector.Vector         `json:"pos"`
	Color      string                `json:"color"`
	Velocity   *vector.Vector        `json:"vel,omitempty"`
	RaysValues *[]float64            `json:"raysValues,omitempty"`
	Brain      *Brain.BrainViewModel `json:"brain,omitempty"`

	LifePoints   int `json:"lifepoints"`
	Energy       int `json:"energy"`
	Reproduction int `json:"reproduction"`
	Digestion    int `json:"digestion"`

	Generation int `json:"generation"`
}

func NewAgentViewModel(agent *Agent, isSelected bool) *AgentViewModel {
	vm := &AgentViewModel{
		ID:       agent.ID,
		Position: agent.Position,
		Color:    agent.Color,
		Velocity: &agent.Velocity,
	}

	if isSelected {

		vm.RaysValues = &agent.RaysValues
		vm.Brain = Brain.NewBrainViewModel(agent.Brain)

		if agent.Color == "Red" {
			vm.LifePoints = (agent.LifePoints * 100) / config.PREDATOR_LIFE_POINTS
			vm.Reproduction = (agent.Reproduction * 100) / config.MAX_REPRODUCTION_PREDATOR
		} else {
			vm.LifePoints = (agent.LifePoints * 100) / config.PREY_LIFE_POINTS
			vm.Reproduction = (agent.Reproduction * 100) / config.MAX_REPRODUCTION_PREY
		}

		if vm.Reproduction > 100 {
			vm.Reproduction = 100
		}

		vm.Energy = (agent.Energy * 100) / config.MAX_ENERGY
		vm.Digestion = agent.Digestion

		vm.Generation = agent.Generation
	}

	return vm
}

func GetGridCell(x, y float64) (uint32, uint32) {
	cellSize := config.GetDefaultConfig().CellSize
	return uint32(x / float64(cellSize)), uint32(y / float64(cellSize))
}

func NewAgent(ID uint32, x, y float64, color string, perceipt Perceipt, brain *Brain.Brain, lifePoint int, generation int) *Agent {
	// random vector of length 1
	vel := vector.Vector{rand.Float64()*2 - 1, rand.Float64()*2 - 1}
	return &Agent{
		ID:         ID,
		Position:   vector.Vector{x, y},
		Color:      color,
		Perceipt:   perceipt,
		RaysValues: make([]float64, config.GetDefaultConfig().RayNumber),
		Brain:      brain,

		LifePoints:   lifePoint,
		Energy:       config.MAX_ENERGY - rand.Intn(50),
		Reproduction: rand.Intn(50),
		Velocity:     vel,

		Generation: generation,
	}
}

func (a *Agent) Move() (oldPosition vector.Vector) {
	speed := a.Speed * float64(config.GetDefaultConfig().MaxSpeed)
	// random angle between 0 and 360 degrees
	rotation := a.Rotation * 360 * 2 * 3.141592653589793

	// Rotate the velocity vector
	a.Velocity = a.Velocity.Rotate(rotation)

	// Scale the velocity by speed
	a.Velocity = a.Velocity.Unit()
	a.Velocity = a.Velocity.Scale(speed)

	oldPosition = a.Position.Clone()
	a.Position = a.Position.Add(a.Velocity)
	if !a.validatePosition() {
		a.Position[0] = a.WrapAround(a.Position[0], float64(config.GetDefaultConfig().Width-1))
		a.Position[1] = a.WrapAround(a.Position[1], float64(config.GetDefaultConfig().Height-1))
	}

	a.CheckCellChange(oldPosition)

	if math.IsNaN(a.Position[0]) {
		fmt.Printf("issue")
	}

	return oldPosition
}

func (a *Agent) validatePosition() bool {
	config := config.GetDefaultConfig()
	return a.Position.X() > 0 && a.Position.X() < float64(config.Width-1) &&
		a.Position.Y() > 0 && a.Position.Y() < float64(config.Width-1)
}

func (a *Agent) CheckCellChange(oldPosition vector.Vector) {
	cellSize := float64(config.GetDefaultConfig().CellSize)
	if (a.Position.X()/cellSize != oldPosition.X()/cellSize) ||
		(a.Position.Y()/cellSize != oldPosition.Y()/cellSize) {
		a.changedCell = true
	}
}

func (a *Agent) ChangedCell() bool {
	return a.changedCell
}

func (a *Agent) ApplyStatsUpdate() (energyLevel int, reproductionLevel int) {
	a.Energy += -(config.ENERGY_LOSS_MULTIPLIER_SPEED*int(a.Speed) + 1)
	if a.Color == "Green" {
		a.Reproduction += config.PREY_REPRODUCTION_GAIN
		if a.Reproduction > config.MAX_REPRODUCTION_PREY {
			a.Reproduction = config.MAX_REPRODUCTION_PREY
		}
	}
	if a.Color == "Red" && a.Digestion > 0 {
		a.Digestion--
	}
	return a.Energy, a.Reproduction
}

func (a *Agent) ApplyDamage(damage int) bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	killed := a.LifePoints > 0 && a.LifePoints-damage <= 0
	a.LifePoints -= damage
	return killed

}
func (a *Agent) WrapAround(value, max float64) float64 {
	if value < 0 {
		return max + math.Mod(value, max)
	}
	return math.Mod(value, max) + 1
}
