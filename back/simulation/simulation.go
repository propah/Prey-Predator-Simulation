package simulation

import (
	"Prey_Predator_MAS/config"
	"Prey_Predator_MAS/environment"
)

type Simulation struct {
	Environment *environment.Environment
}

func NewSimulation() *Simulation {
	config := config.GetDefaultConfig()
	return &Simulation{
		Environment: environment.NewEnvironment(config.Width, config.Height, config.NumAgents, config.CellSize),
	}
}

func (sim *Simulation) Start() {
	sim.Environment.Start()
}
