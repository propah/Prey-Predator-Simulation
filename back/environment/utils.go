package environment

import (
	"Prey_Predator_MAS/config"
	"github.com/quartercastle/vector"
	"math/rand"
)



func generateRandomOffset(color string) vector.Vector {

	if color == "red" {
		return vector.Vector{rand.Float64() * config.AGENT_RADIUS * 10, rand.Float64() * config.AGENT_RADIUS * 10}

	} else {
		return vector.Vector{rand.Float64() * config.AGENT_RADIUS * 15, rand.Float64() * config.AGENT_RADIUS * 15}
	}
}