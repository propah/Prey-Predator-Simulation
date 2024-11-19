package fixedgrid

import (
	"Prey_Predator_MAS/agents"
	"Prey_Predator_MAS/config"
	"sync"

	"github.com/quartercastle/vector"
)

var _ agents.GridAgentProvider = (*FixedGrid)(nil)

type FixedGrid struct {
	rows, cols int
	cellSize   int
	AgentsMap  [config.HEIGHT / config.CELL_SIZE][config.WIDTH / config.CELL_SIZE]AgentStack
	GridMutex  [config.HEIGHT / config.CELL_SIZE][config.WIDTH / config.CELL_SIZE]sync.Mutex
}

func NewFixedGrid() FixedGrid {
	fg := FixedGrid{
		rows:      config.HEIGHT / config.CELL_SIZE,
		cols:      config.WIDTH / config.CELL_SIZE,
		cellSize:  config.CELL_SIZE,
		AgentsMap: [config.HEIGHT / config.CELL_SIZE][config.WIDTH / config.CELL_SIZE]AgentStack{},
		GridMutex: [config.HEIGHT / config.CELL_SIZE][config.WIDTH / config.CELL_SIZE]sync.Mutex{},
	}

	for i := 0; i < config.HEIGHT/config.CELL_SIZE; i++ {
		for j := 0; j < config.WIDTH/config.CELL_SIZE; j++ {
			fg.AgentsMap[i][j] = *NewAgentStack()
		}
	}

	// copy is ok here
	return fg
}

func (fg *FixedGrid) AddAgent(agent *agents.Agent) {
	row, col := agents.GetGridCell(agent.Position.X(), agent.Position.Y())
	fg.GridMutex[row][col].Lock()
	defer fg.GridMutex[row][col].Unlock()
	fg.AgentsMap[row][col].Push(agent)
}

func (fg *FixedGrid) RemoveAgent(agent *agents.Agent, oldPosition vector.Vector) {
	row, col := agents.GetGridCell(oldPosition.X(), oldPosition.Y())
	fg.GridMutex[row][col].Lock()
	defer fg.GridMutex[row][col].Unlock()
	fg.AgentsMap[row][col].Remove(agent)
}
func (fg *FixedGrid) GetAgentsInCell(row, col uint32) []*agents.Agent {
	if row >= uint32(fg.rows) || col >= uint32(fg.cols) || row < 0 || col < 0 {
		return nil
	}

	fg.GridMutex[row][col].Lock()
	defer fg.GridMutex[row][col].Unlock()
	return fg.AgentsMap[row][col].elements[:fg.AgentsMap[row][col].size]
}
