package agents

import (
	"Prey_Predator_MAS/config"
	"fmt"
	"github.com/quartercastle/vector"
	"math"
)

type RayGenerator struct {
	rayNumber int
	rayLength int
	rayAngle  float64
	agentType int // 0 = predator, 1 = prey
}

// generateRays generates the rays for the predator and prey and calculates the bounding box.
func (rg *RayGenerator) generateRays(agent *Agent) (rays []vector.Vector, boundingBox []float64) {
	rays = make([]vector.Vector, rg.rayNumber)

	startAngle := agent.Velocity.Angle() - rg.rayAngle/2
	angleDelta := rg.rayAngle / float64(rg.rayNumber-1)

	minX, minY := agent.Position[0], agent.Position[1]
	maxX, maxY := minX, minY

	for i := 0; i < rg.rayNumber; i++ {

		angle := startAngle + angleDelta*float64(i)

		// Rotate the base vector
		ray := vector.Vector{1, 0}

		cos, sin := math.Cos(angle), math.Sin(angle)

		ax, ay := ray[0], ray[1]
		ray[0] = ax*cos - ay*sin
		ray[1] = ax*sin + ay*cos

		// Scale the ray by the ray length
		// used to be ray = ray.Scale(float64(p.rayLength))
		ray[0] *= float64(rg.rayLength)
		ray[1] *= float64(rg.rayLength)
		rays[i] = ray

		rayEnd := agent.Position.Add(ray)
		minX = math.Min(minX, rayEnd[0])
		maxX = math.Max(maxX, rayEnd[0])
		minY = math.Min(minY, rayEnd.Y())
		maxY = math.Max(maxY, rayEnd.Y())
	}

	return rays, []float64{minX, minY, maxX, maxY}
}

type GridAgentProvider interface {
	GetAgentsInCell(x, y uint32) []*Agent
}

type Perceipt interface {
	Perceive(agent *Agent, grid GridAgentProvider)
}

type PreyPerceipt struct {
	RayGenerator
}

func NewPreyPerceipt(rayNumber, rayLength int, rayAngle float64) Perceipt {
	return &PreyPerceipt{
		RayGenerator: RayGenerator{
			rayNumber: rayNumber,
			rayLength: rayLength,
			rayAngle:  rayAngle * 3.141592653589793 / 180,
			agentType: 1,
		},
	}
}

func (p *PreyPerceipt) Perceive(agent *Agent, grid GridAgentProvider) {
	rays, boundingBox := p.RayGenerator.generateRays(agent)

	evaluatedCells := p.evaluateCellsInFOV(agent, rays, boundingBox)

	gatheredAgents := make([]*Agent, 0, 20)

	for _, cell := range *evaluatedCells {
		agentsInCell := grid.GetAgentsInCell(uint32(cell[0]), uint32(cell.Y()))
		for _, agentInCell := range agentsInCell {
			if agentInCell.ID != agent.ID  && agentInCell.Color != "Green" {
				gatheredAgents = append(gatheredAgents, agentInCell)
			}
		}
	}

	evaluatedCells = nil

	for i := range agent.RaysValues {
		agent.RaysValues[i] = 0
	}

	// check collision between rays and gathered agents
	//fmt.Printf("Gathered agents: %d\n", len(gatheredAgents))
	for _, gatheredAgent := range gatheredAgents {
		for rayIndex, ray := range rays {

			if gatheredAgent != nil && lineCircleCollision(agent.Position[0], agent.Position.Y(), agent.Position[0]+ray[0], agent.Position.Y()+ray.Y(), gatheredAgent.Position[0], gatheredAgent.Position.Y(), float64(config.GetDefaultConfig().AgentRadius*2)) {
				x := gatheredAgent.Position[0] - agent.Position[0]
				y := gatheredAgent.Position[1] - agent.Position[1]
				dist := math.Sqrt(x*x + y*y)

				if dist < agent.RaysValues[rayIndex] || agent.RaysValues[rayIndex] == 0 {
					if gatheredAgent.Color == "Green" {
					agent.RaysValues[rayIndex] = -dist
					} else {
						agent.RaysValues[rayIndex] = dist
					}
				}
			}
		}
	}
}

type PredatorPerceipt struct {
	RayGenerator
}

func NewPredatorPerceipt(rayNumber, rayLength int, rayAngle float64) Perceipt {
	return &PredatorPerceipt{
		RayGenerator: RayGenerator{
			rayNumber: rayNumber,
			rayLength: rayLength,
			rayAngle:  rayAngle * 3.141592653589793 / 180,
			agentType: 0,
		},
	}
}

func (p *PredatorPerceipt) Perceive(agent *Agent, grid GridAgentProvider) {
	rays, boundingBox := p.RayGenerator.generateRays(agent)
	//p.printDebugInfo(agent, rays, boundingBox)

	evaluatedCells := p.evaluateCellsInFOV(agent, rays, boundingBox)
	//p.printEvaluatedCells(evaluatedCells)

	//gather agents in selected cells
	gatheredAgents := make([]*Agent, 0, 20)

	for _, cell := range *evaluatedCells {
		agentsInCell := grid.GetAgentsInCell(uint32(cell[0]), uint32(cell.Y()))
		for _, agentInCell := range agentsInCell {
			if agentInCell.ID != agent.ID && agentInCell.Color != "Red" {
				gatheredAgents = append(gatheredAgents, agentInCell)
			}
		}
	}

	// empty agent.RaysValues
	for i := range agent.RaysValues {
		agent.RaysValues[i] = 0
	}

	// check collision between rays and gathered agents
	//fmt.Printf("Gathered agents: %d\n", len(gatheredAgents))
	for _, gatheredAgent := range gatheredAgents {
		for rayIndex, ray := range rays {
			if gatheredAgent != nil && lineCircleCollision(agent.Position[0], agent.Position.Y(), agent.Position[0]+ray[0], agent.Position.Y()+ray.Y(), gatheredAgent.Position[0], gatheredAgent.Position.Y(), float64(config.GetDefaultConfig().AgentRadius*2)) {
				x := gatheredAgent.Position[0] - agent.Position[0]
				y := gatheredAgent.Position[1] - agent.Position[1]
				dist := math.Sqrt(x*x + y*y)

				if dist < agent.RaysValues[rayIndex] || agent.RaysValues[rayIndex] == 0 {
					if gatheredAgent.Color == "Red" {
						agent.RaysValues[rayIndex] = -dist
					} else {
						agent.RaysValues[rayIndex] = dist
					}
				}
			}
		}
	}
	//fmt.Printf("Agent %d: %v\n", agent.ID, agent.RaysValues)
}

func lineCircleCollision(rayStartX, rayStartY, rayEndX, rayEndY, agentX, agentY, agentRadius float64) bool {
	// Inline pointCircleCollision to avoid function call overhead

	//Verify if the ray' startpoints are inside the circle
	x1 := rayStartX - agentX
	y1 := rayStartY - agentY

	if x1*x1+y1*y1 <= agentRadius*agentRadius {
		return true
	}

	//Verify if the ray' endpoints are inside the circle
	x2 := rayEndX - agentX
	y2 := rayEndY - agentY
	if x2*x2+y2*y2 <= agentRadius*agentRadius {
		return true
	}

	distX := rayEndX - rayStartX
	distY := rayEndY - rayStartY
	lenSq := distX*distX + distY*distY // length squared
	dot := ((agentX-rayStartX)*distX + (agentY-rayStartY)*distY) / lenSq

	closestX := rayStartX + dot*distX
	closestY := rayStartY + dot*distY

	// Check if the closest point is on the line segment
	if !linePointCollision(rayStartX, rayStartY, rayEndX, rayEndY, closestX, closestY) {
		return false
	}

	return pointCircleCollision(closestX, closestY, agentX, agentY, agentRadius)
}

func pointCircleCollision(x, y, agentX, agentY, agentRadius float64) bool {
	x1 := x - agentX
	y1 := y - agentY
	return x1*x1+y1*y1 <= agentRadius*agentRadius
}

func linePointCollision(rayStartX, rayStartY, rayEndX, rayEndY, pointX, pointY float64) bool {
	// Compute the squared lengths
	lenSq := (rayEndX-rayStartX)*(rayEndX-rayStartX) + (rayEndY-rayStartY)*(rayEndY-rayStartY)
	d1Sq := (pointX-rayStartX)*(pointX-rayStartX) + (pointY-rayStartY)*(pointY-rayStartY)
	d2Sq := (pointX-rayEndX)*(pointX-rayEndX) + (pointY-rayEndY)*(pointY-rayEndY)

	if d1Sq == 0 || d2Sq == 0 {
		return true
	}
	if d1Sq > lenSq || d2Sq > lenSq {
		return false
	}
	dot := (pointX-rayStartX)*(rayEndX-rayStartX) + (pointY-rayStartY)*(rayEndY-rayStartY)
	if dot < 0 || dot > lenSq {
		return false
	}
	crossProductSq := d1Sq - dot*dot/lenSq

	toleranceSq := 0.00001 * 0.00001

	return crossProductSq <= toleranceSq
}

// printDebugInfo prints debugging information about the agent and rays.
func (p *PredatorPerceipt) printDebugInfo(agent *Agent, rays []vector.Vector, boundingBoxX []float64) {
	fmt.Printf("Agent Position: [%f,%f]\n", agent.Position[0], agent.Position.Y())
	fmt.Printf("Bounding Box: minX=%f, maxX=%f, minY=%f, maxY=%f\n", boundingBoxX[0], boundingBoxX[2], boundingBoxX[1], boundingBoxX[3])
	fmt.Printf("Agent Velocity: [%f,%f]\n", agent.Velocity[0], agent.Velocity.Y())
	for _, ray := range rays {
		fmt.Printf("Ray: [%f,%f]\n", ray[0], ray.Y())
	}
}

// printEvaluatedCells prints the evaluated cells.
func (p *PredatorPerceipt) printEvaluatedCells(cells []vector.Vector) {
	cellSize := float64(config.GetDefaultConfig().CellSize)
	for _, cell := range cells {
		fmt.Printf("Cell: Indexes: [%f,%f] | Coords: [%f,%f]\n", cell[0], cell.Y(), cell[0]*float64(cellSize), cell.Y()*float64(cellSize))
	}
}

// alignToGrid aligns the given coordinates to the grid.
func (rg *RayGenerator) alignToGrid(x, y float64) (float64, float64) {
	cellSize := config.GetDefaultConfig().CellSize
	// Convert to integer for bitwise operation
	xi, yi := int(x), int(y)

	// Perform bitwise AND with CELL_SIZE - 1
	alignedX := xi & ^(cellSize - 1)
	alignedY := yi & ^(cellSize - 1)

	// Convert back to float64
	return float64(alignedX), float64(alignedY)
}

// cellInTriangle checks if any part of the cell is within the predator's triangle of vision.
func (rg *RayGenerator) cellInTriangle(x, y float64, agent *Agent, rays []vector.Vector) bool {
	cellSize := float64(config.GetDefaultConfig().CellSize)
	points := []vector.Vector{
		{x, y},
		{x + cellSize, y},
		{x, y + cellSize},
		{x + cellSize, y + cellSize},
		{x + cellSize/2, y + cellSize/2},
	}
	triangle := []vector.Vector{agent.Position, agent.Position.Add(rays[0]), agent.Position.Add(rays[len(rays)-1])}
	for _, pt := range points {
		if pointInTriangle(pt, triangle[0], triangle[1], triangle[2]) {
			return true
		}
	}
	return false
}

// evaluateCellsInFOV evaluates which cells fall within the field of view of the predator.
func (rg *RayGenerator) evaluateCellsInFOV(agent *Agent, rays []vector.Vector, boundingBox []float64) *[]vector.Vector {
	config := config.GetDefaultConfig()
	minX := math.Max(0, math.Min(boundingBox[0], float64(config.Width-1)))
	maxX := math.Max(0, math.Min(boundingBox[2], float64(config.Width-1)))
	minY := math.Max(0, math.Min(boundingBox[1], float64(config.Height-1)))
	maxY := math.Max(0, math.Min(boundingBox[3], float64(config.Height-1)))

	firstCellX, firstCellY := rg.alignToGrid(minX, minY)
	lastCellX, lastCellY := rg.alignToGrid(maxX, maxY)

	if firstCellX < 0 || firstCellY < 0 || lastCellX < 0 || lastCellY < 0 {
		fmt.Printf("ERROR: Negative cell coordinates: [%f,%f] [%f,%f]\n", firstCellX, firstCellY, lastCellX, lastCellY)
	}
	evaluatedCells := make([]vector.Vector, 0, 10)
	evaluatedCells = append(evaluatedCells, vector.Vector{math.Floor(agent.Position[0] / float64(config.CellSize)), math.Floor(agent.Position.Y() / float64(config.CellSize))})
	for x := firstCellX; x <= lastCellX; x += float64(config.CellSize) {
		for y := firstCellY; y <= lastCellY; y += float64(config.CellSize) {
			if rg.agentType == 0 && rg.cellInTriangle(x, y, agent, rays) {
				evaluatedCells = append(evaluatedCells, vector.Vector{x / float64(config.CellSize), y / float64(config.CellSize)})
			} else if rg.agentType == 1 {
				evaluatedCells = append(evaluatedCells, vector.Vector{x / float64(config.CellSize), y / float64(config.CellSize)})
			}
		}
	}
	return &evaluatedCells
}

func pointInTriangle(pt, v1, v2, v3 vector.Vector) bool {
	// Barycentric coordinates
	d1 := sign(pt, v1, v2)
	d2 := sign(pt, v2, v3)
	d3 := sign(pt, v3, v1)

	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)

	return !(hasNeg && hasPos)
}

func sign(p1, p2, p3 vector.Vector) float64 {
	return (p1[0]-p3[0])*(p2.Y()-p3.Y()) - (p2[0]-p3[0])*(p1.Y()-p3.Y())
}
