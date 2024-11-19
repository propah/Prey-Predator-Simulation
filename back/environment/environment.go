package environment

import (
	"Prey_Predator_MAS/Brain"
	"Prey_Predator_MAS/agents"
	"Prey_Predator_MAS/config"
	"Prey_Predator_MAS/fixedgrid"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Environment struct {
	Width, Height    int
	Agents           []*agents.Agent
	IterationDone    chan bool
	wg               sync.WaitGroup
	fixedGrid        fixedgrid.FixedGrid
	predatorPerceipt agents.Perceipt
	preyPerceipt     agents.Perceipt
	lock             sync.Mutex
	steps            int
	PreyCount        int
	PredatorCount    int
	newAgents        []*agents.Agent
	idCounter        uint32
	TickCounter      uint64
	StartTime        time.Time
}

func NewEnvironment(width, height, numAgents, fixedGridCellSize int) *Environment {
	config := config.GetDefaultConfig()
	env := &Environment{
		Width:            width,
		Height:           height,
		Agents:           make([]*agents.Agent, 0),
		IterationDone:    make(chan bool),
		fixedGrid:        fixedgrid.NewFixedGrid(),
		predatorPerceipt: agents.NewPredatorPerceipt(config.RayNumber, config.PredatorRayLength, float64(config.PredatorRayAngleDeg)),
		preyPerceipt:     agents.NewPreyPerceipt(config.RayNumber, config.PreyRayLength, float64(config.PreyRayAngleDeg)),
		PreyCount:        0,
		PredatorCount:    0,
		newAgents:        make([]*agents.Agent, 0),
		idCounter:        1,
		TickCounter:      0,
		StartTime:        time.Now(),
	}

	for i := 0; i < numAgents; i++ {
		// init agents
		var agentColor string
		var perceipt agents.Perceipt
		var lifePoints int
		if i%2 == 0 {
			agentColor = "Red"
			perceipt = env.predatorPerceipt
			lifePoints = config.PredatorLifePoints
			env.PredatorCount++

		} else {
			agentColor = "Green"
			perceipt = env.preyPerceipt
			lifePoints = config.PreyLifePoints
			env.PreyCount++
		}
		var x, y float64
		x = float64(rand.Intn(width - 1))
		y = float64(rand.Intn(height - 1))

		brain := Brain.NewBrain(config.InputNeuronNumber, config.OutputNeuronNumber)

		env.Agents = append(env.Agents, agents.NewAgent(uint32(env.idCounter), x, y, agentColor, perceipt, brain, lifePoints, 1))

		env.idCounter++
		// add agent to fixed grid
		env.fixedGrid.AddAgent(env.Agents[i])
	}
	return env
}

func (e *Environment) Start() {
	for {
		start := time.Now()

		// Perception phase
		e.wg.Add(len(e.Agents))
		for _, agent := range e.Agents {
			if math.IsNaN(agent.Position[0]) {
				fmt.Printf("issue")
			}
			go func(agent *agents.Agent, fixedGrid *fixedgrid.FixedGrid) {
				defer e.wg.Done()
				agent.Perceipt.Perceive(agent, fixedGrid)
			}(agent, &e.fixedGrid)
		}
		e.wg.Wait() // Wait for all agents to complete the perception phase

		// think phase
		e.wg.Add(len(e.Agents))

		for _, agent := range e.Agents {
			go func(agent *agents.Agent) {
				defer e.wg.Done()
				//agent.Brain.Mutate()
				agent.Speed, agent.Rotation = agent.Brain.TakeDecision(agent.RaysValues, agent.Color)
				if agent.Speed > 1 {
					agent.Speed = 1
				} else if agent.Speed < 0 {
					agent.Speed = 0
				}
			}(agent)
		}
		e.wg.Wait() // Wait for all agents to complete the perception phase

		// Action phase
		e.wg.Add(len(e.Agents))
		for index, agent := range e.Agents {
			go func(agent *agents.Agent, index int) {
				defer e.wg.Done()
				if !agent.Regen {
					if agent.Color == "Red" && e.steps < 1600 {
						agent.Reproduction += config.MAX_REPRODUCTION_PREDATOR / 90
						if agent.Reproduction > config.MAX_REPRODUCTION_PREDATOR {
							agent.Reproduction = config.MAX_REPRODUCTION_PREDATOR
						}
					}

					oldPos := agent.Move()
					if agent.ChangedCell() {
						e.fixedGrid.RemoveAgent(agent, oldPos)
						e.fixedGrid.AddAgent(agent)
					}

					energy, _ := agent.ApplyStatsUpdate()
					e.lock.Lock()
					if ((agent.Reproduction >= config.MAX_REPRODUCTION_PREDATOR && agent.Color == "Red") || (agent.Reproduction >= config.MAX_REPRODUCTION_PREY && agent.Color == "Green")) && ((agent.Color == "Red" && e.PredatorCount < config.MAX_PREDATOR) ||
						(agent.Color == "Green" && e.PreyCount < config.MAX_PREY)) {
						// reproduction

						randomOffset := generateRandomOffset(agent.Color)

						var x, y float64
						x = agent.Position.X() + randomOffset[0]
						y = agent.Position.Y() + randomOffset[1]

						// constrain x and y to be modulo width and height
						x = agent.WrapAround(x, float64(e.Width-1))
						y = agent.WrapAround(y, float64(e.Height-1))

						brain := agent.Brain.Copy()
						for i := 0; i < 1; i++ {
							brain.Mutate()
						}

						generation := agent.Generation + 1
						newAgent := agents.NewAgent(e.idCounter, x, y, agent.Color, agent.Perceipt, brain, agent.LifePoints, generation)
						e.idCounter++
						e.Agents = append(e.Agents, newAgent)
						e.fixedGrid.AddAgent(newAgent)
						if agent.Color == "Red" {
							e.PredatorCount++
						} else {
							e.PreyCount++
						}

						agent.Reproduction = 0
					}
					e.HandleAgentCollision(agent)
					e.lock.Unlock()

					if agent.Color == "Red" && energy <= 0 {
						agent.LifePoints = 0
					} else if agent.Color == "Green" && energy <= 0 {
						agent.Regen = true
					}
				} else {
					if agent.Energy >= config.MAX_ENERGY {
						agent.Regen = false
						agent.Energy = config.MAX_ENERGY
					} else {
						agent.Energy += config.PREY_ENERGY_GAIN
					}
				}
				if math.IsNaN(agent.Position[0]) {
					fmt.Printf("issue")
				}

			}(agent, index)
		}
		e.wg.Wait() // Wait for all agents to complete the action phase
		e.removeDeadAgents()
		e.TickCounter++

		elapsed := time.Since(start)

		// limit at 60 updates per second
		time.Sleep(time.Second/60 - elapsed)
		fmt.Printf("Iteration took %dms - tickNb: %d - agentNb: %d\n", elapsed.Milliseconds(), e.TickCounter, len(e.Agents))
		e.IterationDone <- true
		e.steps++
	}
}

func (e *Environment) HandleAgentCollision(agent *agents.Agent) {
	// get cells around agent
	agents := make([]*agents.Agent, 0, 40)
	for i := uint32(0); i < 3; i++ {
		for j := uint32(0); j < 3; j++ {
			agents = append(agents, e.fixedGrid.GetAgentsInCell(uint32(agent.Position.X()/float64(config.GetDefaultConfig().CellSize))-1+i, uint32(agent.Position.Y()/float64(config.GetDefaultConfig().CellSize))-1+j)...)
		}
	}

	for _, otherAgent := range agents {
		if otherAgent != nil && otherAgent != agent && otherAgent.LifePoints > 0 {
			x := otherAgent.Position[0] - agent.Position[0]
			y := otherAgent.Position[1] - agent.Position[1]
			distSquared := x*x + y*y
			radiusSquared := float64(config.AGENT_RADIUS * 2 * config.AGENT_RADIUS * 2 * 4)

			if distSquared < radiusSquared {
				if agent.Color == otherAgent.Color {

					/*dist := math.Sqrt(distSquared)
					separationDist := config.AGENT_RADIUS - dist/2

					x /= dist
					y /= dist

					oldPos1 := agent.Position
					oldPos2 := otherAgent.Position

					agent.Position[0] -= x * separationDist
					agent.Position[1] -= y * separationDist
					otherAgent.Position[0] += x * separationDist
					otherAgent.Position[1] += y * separationDist

					agent.Position[0] = math.Mod(agent.Position[0], float64(e.Width-1)) + 1
					agent.Position[1] = math.Mod(agent.Position[1], float64(e.Height-1)) + 1
					otherAgent.Position[0] = math.Mod(otherAgent.Position[0], float64(e.Width-1)) + 1
					otherAgent.Position[1] = math.Mod(otherAgent.Position[1], float64(e.Height-1)) + 1
					*/
					// Calculate overlap
					/*
						oldPos1 := agent.Position.Clone()
						oldPos2 := otherAgent.Position.Clone()

						dist := math.Sqrt(distSquared)
						pushDistance := float64(config.AGENT_RADIUS)

						pushX, pushY := x/dist+0.001, y/dist+0.001

						if (agent.Velocity[0] == 0 && agent.Velocity[1] == 0) ||
							(otherAgent.Velocity[0] == 0 && otherAgent.Velocity[1] == 0) {
							pushX, pushY = randomDirection()
						}

						// Move the agents apart by the push distance
						agent.Position[0] -= pushX * pushDistance
						agent.Position[1] -= pushY * pushDistance
						otherAgent.Position[0] += pushX * pushDistance
						otherAgent.Position[1] += pushY * pushDistance

						// Clamp positions to environment bounds
						agent.Position[0] = math.Mod(agent.Position[0], float64(e.Width-1)) + 1
						agent.Position[1] = math.Mod(agent.Position[1], float64(e.Height-1)) + 1
						otherAgent.Position[0] = math.Mod(otherAgent.Position[0], float64(e.Width-1)) + 1
						otherAgent.Position[1] = math.Mod(otherAgent.Position[1], float64(e.Height-1)) + 1

						agent.CheckCellChange(oldPos1)
						otherAgent.CheckCellChange(oldPos2)
					*/
					if math.IsNaN(agent.Position[0]) {
						fmt.Printf("issue")
					}
				}

				if agent.Color == "Red" && otherAgent.Color == "Green" {
					killed := otherAgent.ApplyDamage(config.PREDATOR_ATTACK_DAMAGE)
					if killed && agent.Digestion == 0 {
						agent.Energy += config.PREDATOR_ENERGY_GAIN
						if agent.Energy > config.MAX_ENERGY {
							agent.Energy = config.MAX_ENERGY
						}
						agent.Digestion = 10
						agent.Reproduction += config.PREDATOR_REPRODUCTION_GAIN
						e.fixedGrid.RemoveAgent(otherAgent, otherAgent.Position)
						//e.PreyCount--
					}
				} else if agent.Color == "Green" && otherAgent.Color == "Red" {
					/*
						if otherAgent.ApplyDamage(config.PREY_ATTACK_DAMAGE) {
							e.fixedGrid.RemoveAgent(otherAgent, otherAgent.Position)
						}

					*/
				}
			}
		}
	}
}

func randomDirection() (float64, float64) {
	angle := rand.Float64() * 2 * math.Pi   // Random angle in radians
	return math.Cos(angle), math.Sin(angle) // Return the x and y components of the direction
}

func (e *Environment) removeDeadAgents() {
	aliveAgents := make([]*agents.Agent, 0, len(e.Agents)) // Preallocate space for efficiency
	for _, agent := range e.Agents {
		if agent.LifePoints > 0 { // Keep only non-nil agents
			aliveAgents = append(aliveAgents, agent)
		} else {
			e.fixedGrid.RemoveAgent(agent, agent.Position)
			if agent.Color == "Red" {
				e.PredatorCount--
			} else {
				e.PreyCount--
			}
		}
	}
	e.Agents = aliveAgents
}

func (e *Environment) LongPollIterationEnd() {
	<-e.IterationDone
}
