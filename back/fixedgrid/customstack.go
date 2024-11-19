package fixedgrid

import (
	"Prey_Predator_MAS/agents"
	"Prey_Predator_MAS/config"
	"fmt"
)

type AgentStack struct {
	elements 					[config.APPROXIMATE_MAX_AGENT_PER_CELL]*agents.Agent
	size	 					int
}

func NewAgentStack() *AgentStack {
	return &AgentStack{
		elements:	[config.APPROXIMATE_MAX_AGENT_PER_CELL]*agents.Agent{},
		size:     	0,
	}
}

func (s *AgentStack) Push(agent *agents.Agent) {
	if s.size >= config.APPROXIMATE_MAX_AGENT_PER_CELL {
		fmt.Println("ERROR: AgentStack is full, need to increase APPROXIMATE_MAX_AGENT_PER_CELL")
		return
	}
	s.elements[s.size] = agent
	s.size++
}

func (s *AgentStack) Remove(agent *agents.Agent) {
	for i := 0; i < s.size; i++ {
		if s.elements[i] == agent {
			s.elements[i] = s.elements[s.size-1]
			s.elements[s.size-1] = nil
			s.size--
			return
		}
	}
}
