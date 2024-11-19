package Brain

import (
	"Prey_Predator_MAS/config"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

type Brain struct {
	InputNeurons  []*Neuron
	OutputNeurons []*Neuron
	HiddenNeurons []*Neuron
	Connections   []*Connection
	finalDepth    int
}

type Neuron struct {
	Value       float64
	Bias        float64
	Connections []*Connection
	Depth       int
}

type Connection struct {
	Source *Neuron
	Target *Neuron
	Weight float64
}

func NewBrain(numInputs, numOutputs int) *Brain {
	brain := &Brain{
		InputNeurons:  make([]*Neuron, numInputs+1),
		OutputNeurons: make([]*Neuron, numOutputs),
		HiddenNeurons: make([]*Neuron, 0, 10),
		Connections:   make([]*Connection, 0, 30),
		finalDepth:    1,
	}

	for i := 0; i < numInputs+1; i++ {
		brain.InputNeurons[i] = &Neuron{
			Value:       0,
			Bias:        0,
			Connections: make([]*Connection, 0, 10),
			Depth:       0,
		}
	}

	for i := 0; i < numOutputs; i++ {
		brain.OutputNeurons[i] = &Neuron{
			Value:       0,
			Bias:        0,
			Connections: make([]*Connection, 0, 10),
			Depth:       1,
		}
	}

	for i := 0; i < config.START_MUTATION_NUMBER; i++ {
		brain.Mutate()
	}

	return brain
}

func (b *Brain) Mutate() {
	mutationRates := config.GetDefaultMutationRate()

	// Handle edge case and adapt mutations rates accordingly
	if len(b.Connections) == 0 {
		mutationRates.WeightMutationRate = 0
		mutationRates.DelConnectionRate = 0
		mutationRates.NewNeuronRate = 0
	}

	if len(b.HiddenNeurons) == 0 {
		mutationRates.DelNeuronRate = 0
	}

	if len(b.HiddenNeurons) >= config.MAX_NEURON_NUMBER {
		mutationRates.NewNeuronRate = 0
	}

	chosenMutation := weightedRandom(mutationRates)
	if chosenMutation == -1 {
		fmt.Printf("ERROR: No mutation was chosen\n")
		return
	}

	// 0 = no mutation
	// 1 = weight mutation
	// 2 = bias mutation
	// 3 = new connection
	// 4 = del connection
	// 5 = new neuron
	// 6 = del neuron

	switch chosenMutation {
	case 0:
		// No mutation
		return
	case 1:
		// Weight mutation
		b.weightMutation()
	case 2:
		// Bias mutation
		b.biasMutation()
	case 3:
		// New connection
		b.newConnection()
	case 4:
		// Del connection
		b.delConnection()
	case 5:
		// New neuron
		b.newNeuron()
	case 6:
		// Del neuron
		b.delNeuron()
	}

}

func (b *Brain) weightMutation() {
	randomConIndex := rand.Intn(len(b.Connections))
	randomCon := b.Connections[randomConIndex]

	change := rand.NormFloat64() * config.WEIGHT_MUTATION_STAND_DEV
	randomCon.Weight += change

	//fmt.Printf("Connection %d weight changed by %f\n", randomConIndex, change)
}

func (b *Brain) biasMutation() {
	randomNeuronIndex := rand.Intn(len(b.HiddenNeurons) + len(b.OutputNeurons))
	if randomNeuronIndex < len(b.OutputNeurons) {
		b.OutputNeurons[randomNeuronIndex].Bias += rand.NormFloat64() * config.BIAS_MUTATION_STAND_DEV
	} else {
		b.HiddenNeurons[randomNeuronIndex-len(b.OutputNeurons)].Bias += rand.NormFloat64() * config.BIAS_MUTATION_STAND_DEV
	}

	//fmt.Printf("Neuron %d bias changed\n", randomNeuronIndex)
}

func (b *Brain) newConnection() {
	// Select a random source neuron
	randomSourceNeuronIndex := rand.Intn(len(b.InputNeurons) + len(b.HiddenNeurons))
	var sourceNeuron *Neuron
	if randomSourceNeuronIndex < len(b.InputNeurons) {
		sourceNeuron = b.InputNeurons[randomSourceNeuronIndex]
	} else {
		sourceNeuron = b.HiddenNeurons[randomSourceNeuronIndex-len(b.InputNeurons)]
	}

	// Filter neurons to find those with a greater depth than the source neuron
	var targetNeurons []*Neuron
	for _, neuron := range b.HiddenNeurons {
		if neuron.Depth > sourceNeuron.Depth {
			targetNeurons = append(targetNeurons, neuron)
		}
	}
	for _, neuron := range b.OutputNeurons {
		if neuron.Depth > sourceNeuron.Depth {
			targetNeurons = append(targetNeurons, neuron)
		}
	}

	var availableTargetNeurons []*Neuron
    for _, neuron := range targetNeurons {
        if !b.connectionExists(sourceNeuron, neuron) {
            availableTargetNeurons = append(availableTargetNeurons, neuron)
        }
    }

	// Check if there are any available target neurons
	if len(availableTargetNeurons) == 0 {
		return // No valid target neuron, so exit the function
	}

	// Select a random target neuron from this filtered subset

	randomTargetNeuronIndex := rand.Intn(len(availableTargetNeurons))
    targetNeuron := availableTargetNeurons[randomTargetNeuronIndex]

	// Check if connection already exists
	// for _, connection := range sourceNeuron.Connections {
	// 	if connection.Target == targetNeuron {
	// 		// Strengthens connection
	// 		connection.Weight += math.Abs(rand.NormFloat64() * config.WEIGHT_MUTATION_STAND_DEV)
	// 	}
	// }

	// Create a new connection with an initial weight and add it to Connections
	newConnection := &Connection{
		Source: sourceNeuron,
		Target: targetNeuron,
		Weight: rand.NormFloat64() * config.WEIGHT_MUTATION_STAND_DEV,
	}
	b.Connections = append(b.Connections, newConnection)

	// update source neuron's connections
	sourceNeuron.Connections = append(sourceNeuron.Connections, newConnection)

	// Update the depth of subsequent neurons recursively
	b.updateDepth(targetNeuron, sourceNeuron.Depth+1)

	//fmt.Printf("New connection created from neuron %d to neuron %d - Depht %d - %d\n", randomSourceNeuronIndex, randomTargetNeuronIndex, sourceNeuron.Depth, targetNeuron.Depth)
}

func (b *Brain) delConnection() {
	randomConIndex := rand.Intn(len(b.Connections))
	randomCon := b.Connections[randomConIndex]

	// remove connection from source neuron
	for i, con := range randomCon.Source.Connections {
		if con == randomCon {
			randomCon.Source.Connections = append(randomCon.Source.Connections[:i], randomCon.Source.Connections[i+1:]...)
			break
		}
	}
	// remove connection from brain
	b.Connections = append(b.Connections[:randomConIndex], b.Connections[randomConIndex+1:]...)

	// check depth
	for _, neuron := range b.InputNeurons {
		b.updateDepth(neuron, 0)
	}

	//fmt.Printf("Connection %d deleted\n", randomConIndex)
}

func (b *Brain) newNeuron() {
	// Select a random connection
	randomConIndex := rand.Intn(len(b.Connections))
	randomCon := b.Connections[randomConIndex]

	// instantiate new neuron
	newNeuron := &Neuron{
		Value:       0,
		Bias:        rand.NormFloat64() * config.BIAS_MUTATION_STAND_DEV,
		Connections: make([]*Connection, 0, 10),
		Depth:       randomCon.Source.Depth + 1,
	}

	// create new connection from source to new neuron
	newCon1 := &Connection{
		Source: randomCon.Source,
		Target: newNeuron,
		Weight: 1,
	}

	// create new connection from new neuron to target
	newCon2 := &Connection{
		Source: newNeuron,
		Target: randomCon.Target,
		Weight: randomCon.Weight,
	}

	// remove old connection
	for i, con := range randomCon.Source.Connections {
		if con == randomCon {
			randomCon.Source.Connections = append(randomCon.Source.Connections[:i], randomCon.Source.Connections[i+1:]...)
			break
		}
	}

	// remove old connection from brain
	b.Connections = append(b.Connections[:randomConIndex], b.Connections[randomConIndex+1:]...)

	// add new neuron to brain
	b.HiddenNeurons = append(b.HiddenNeurons, newNeuron)

	// add new connections to brain
	b.Connections = append(b.Connections, newCon1, newCon2)

	// update source neuron's connections
	randomCon.Source.Connections = append(randomCon.Source.Connections, newCon1)

	// update new neuron's connections
	newNeuron.Connections = append(newNeuron.Connections, newCon2)

	// Update the depth of subsequent neurons recursively
	b.updateDepth(randomCon.Target, newNeuron.Depth+1)

	//fmt.Printf("New neuron created between neuron %d and neuron %d\n", randomCon.Source, randomCon.Target)
}

func (b *Brain) delNeuron() {
	// Select a random neuron
	randomNeuronIndex := rand.Intn(len(b.HiddenNeurons))
	randomNeuron := b.HiddenNeurons[randomNeuronIndex]

	// remove neuron from brain
	b.HiddenNeurons = append(b.HiddenNeurons[:randomNeuronIndex], b.HiddenNeurons[randomNeuronIndex+1:]...)

	// Iterate in reverse to avoid index issues when removing connections
	for conIndex := len(b.Connections) - 1; conIndex >= 0; conIndex-- {
		con := b.Connections[conIndex]
		if con.Source == randomNeuron || con.Target == randomNeuron {
			if con.Target == randomNeuron {
				// remove the connection from the source neuron
				for i, sourceCon := range con.Source.Connections {
					if con == sourceCon {
						con.Source.Connections = append(con.Source.Connections[:i], con.Source.Connections[i+1:]...)
						break
					}
				}
			}

			// remove the connection from the brain
			b.Connections = append(b.Connections[:conIndex], b.Connections[conIndex+1:]...)

		}
	}

	for _, neuron := range b.InputNeurons {
		b.updateDepth(neuron, 0)
	}

	//fmt.Printf("Neuron %d deleted\n", randomNeuronIndex)
}

func (b *Brain) TakeDecision(input []float64, clr string) (speed, rotation float64) {
	// reset hidden and output neurons
	for _, neuron := range b.HiddenNeurons {
		neuron.Value = 0
	}

	for _, neuron := range b.OutputNeurons {
		neuron.Value = 0
	}

	var scaleValue float64
	if clr == "Red" {
		scaleValue = config.PREDATOR_RAY_LENGTH
	} else {
		scaleValue = config.PREY_RAY_LENGTH
	}

	// Set input neurons values
	for i := 0; i < len(b.InputNeurons)-1; i += 1 {
		b.InputNeurons[i].Value = math.Abs(input[i]) / scaleValue

		// add the value of the input neuron to the value of the connected neurons
		for _, connection := range b.InputNeurons[i].Connections {
			if (connection.Source != nil) && (connection.Target != nil) {
				connection.Target.Value += b.InputNeurons[i].Value * connection.Weight
			} else {
				fmt.Printf("ERROR: NIL target in input neuron of brain: %p\n", b)
				return
			}
		}
		// inputType := input[i+1]
		// if inputType < 0 {
		// 	inputType = -1
		// } else if inputType > 0 {
		// 	inputType = 1
		// }
		// b.InputNeurons[i+1].Value = input[i+1]

		// // add the value of the input neuron to the value of the connected neurons
		// for _, connection := range b.InputNeurons[i+1].Connections {
		// 	if (connection.Source != nil) && (connection.Target != nil) {
		// 		connection.Target.Value += b.InputNeurons[i+1].Value * connection.Weight
		// 	} else {
		// 		fmt.Printf("ERROR: NIL target in input neuron of brain: %p\n", b)
		// 		return
		// 	}
		// }

	}

	// Set bias neuron value
	b.InputNeurons[len(b.InputNeurons)-1].Value = 1

	// Set hidden neurons values
	for _, neuron := range b.HiddenNeurons {
		// apply bias
		neuron.Value += neuron.Bias

		// apply activation function
		neuron.Value = sigmoid(neuron.Value)

		// add the value of the hidden neuron to the value of the connected neurons
		for _, connection := range neuron.Connections {
			connection.Target.Value += neuron.Value * connection.Weight
		}
	}

	// Set output neurons values
	b.OutputNeurons[0].Value += b.OutputNeurons[0].Bias
	b.OutputNeurons[1].Value += b.OutputNeurons[1].Bias

	// No activation function for speed

	b.OutputNeurons[1].Value = tanh(b.OutputNeurons[1].Value)
	b.OutputNeurons[1].Value *= 3.141592653589793
	// return output neurons values
	output := make([]float64, len(b.OutputNeurons))
	for i, neuron := range b.OutputNeurons {
		output[i] = neuron.Value
	}

	//b.printBrainState()

	return output[0], output[1]
}

func (b *Brain) printBrainState() {
	fmt.Printf("Input neurons:\n")
	for i, neuron := range b.InputNeurons {
		fmt.Printf("Neuron %d: Value: %f, Bias: %f\n", i, neuron.Value, neuron.Bias)
		for _, connection := range neuron.Connections {
			fmt.Printf("Connection to neuron %d: Weight: %f\n", connection.Target, connection.Weight)
		}
	}
	fmt.Printf("Hidden neurons:\n")
	for i, neuron := range b.HiddenNeurons {
		fmt.Printf("Neuron %d: Value: %f, Bias: %f\n", i, neuron.Value, neuron.Bias)
		for _, connection := range neuron.Connections {
			fmt.Printf("Connection to neuron %d: Weight: %f\n", connection.Target, connection.Weight)
		}
	}
	fmt.Printf("Output neurons:\n")
	for i, neuron := range b.OutputNeurons {
		fmt.Printf("Neuron %d: Value: %f, Bias: %f\n", i, neuron.Value, neuron.Bias)
		for _, connection := range neuron.Connections {
			fmt.Printf("Connection to neuron %d: Weight: %f\n", connection.Target, connection.Weight)
		}
	}
}

func (b *Brain) updateDepth(neuron *Neuron, newDepth int) {
	if neuron.Depth < newDepth {
		neuron.Depth = newDepth

		for _, connection := range neuron.Connections {
			b.updateDepth(connection.Target, newDepth+1)
		}
	}

	// sort the hidden neurons by depth
	sort.Slice(b.HiddenNeurons, func(i, j int) bool {
		return b.HiddenNeurons[i].Depth < b.HiddenNeurons[j].Depth
	})

	if len(b.HiddenNeurons) > 0 && b.HiddenNeurons[len(b.HiddenNeurons)-1].Depth >= b.finalDepth {
		b.finalDepth = b.HiddenNeurons[len(b.HiddenNeurons)-1].Depth
		b.OutputNeurons[0].Depth = b.finalDepth + 1
		b.OutputNeurons[1].Depth = b.finalDepth + 1
	}
}

func sigmoid(x float64) float64 {
	//return 1 / (1 + math.Exp(-x))
	return x
}

func tanh(x float64) float64 {
	return math.Tanh(x)
}
func weightedRandom(mutationRates config.MutationRate) int {
	totalWeight := mutationRates.NoMutation +
		mutationRates.WeightMutationRate +
		mutationRates.BiasMutationRate +
		mutationRates.NewConnectionRate +
		mutationRates.DelConnectionRate +
		mutationRates.NewNeuronRate +
		mutationRates.DelNeuronRate

	ranNum := rand.Intn(totalWeight)

	if ranNum < mutationRates.NoMutation {
		return 0
	}
	ranNum -= mutationRates.NoMutation

	if ranNum < mutationRates.WeightMutationRate {
		return 1
	}
	ranNum -= mutationRates.WeightMutationRate

	if ranNum < mutationRates.BiasMutationRate {
		return 2
	}
	ranNum -= mutationRates.BiasMutationRate

	if ranNum < mutationRates.NewConnectionRate {
		return 3
	}
	ranNum -= mutationRates.NewConnectionRate

	if ranNum < mutationRates.DelConnectionRate {
		return 4
	}
	ranNum -= mutationRates.DelConnectionRate

	if ranNum < mutationRates.NewNeuronRate {
		return 5
	}

	return 6

}

func (b *Brain) Copy() *Brain {
	neuronMap := make(map[*Neuron]*Neuron) // Map to track old to new neuron mapping

	// Create new Brain instance
	brain := &Brain{
		InputNeurons:  make([]*Neuron, len(b.InputNeurons)),
		OutputNeurons: make([]*Neuron, len(b.OutputNeurons)),
		HiddenNeurons: make([]*Neuron, len(b.HiddenNeurons)),
		Connections:   make([]*Connection, len(b.Connections)),
		finalDepth:    b.finalDepth,
	}

	// Copy input neurons and fill the map
	for i, oldNeuron := range b.InputNeurons {
		newNeuron := &Neuron{
			Value:       oldNeuron.Value,
			Bias:        oldNeuron.Bias,
			Connections: make([]*Connection, 0, len(oldNeuron.Connections)), // Initially empty
			Depth:       oldNeuron.Depth,
		}
		neuronMap[oldNeuron] = newNeuron
		brain.InputNeurons[i] = newNeuron
	}

	// Copy hidden neurons
	for i, oldNeuron := range b.HiddenNeurons {
		newNeuron := &Neuron{
			Value:       oldNeuron.Value,
			Bias:        oldNeuron.Bias,
			Connections: make([]*Connection, 0, len(oldNeuron.Connections)),
			Depth:       oldNeuron.Depth,
		}
		neuronMap[oldNeuron] = newNeuron
		brain.HiddenNeurons[i] = newNeuron
	}

	// Copy output neurons
	for i, oldNeuron := range b.OutputNeurons {
		newNeuron := &Neuron{
			Value:       oldNeuron.Value,
			Bias:        oldNeuron.Bias,
			Connections: make([]*Connection, 0, len(oldNeuron.Connections)),
			Depth:       oldNeuron.Depth,
		}
		neuronMap[oldNeuron] = newNeuron
		brain.OutputNeurons[i] = newNeuron
	}

	// Copy connections and update neuron references
	for i, oldConnection := range b.Connections {
		newConnection := &Connection{
			Source: neuronMap[oldConnection.Source],
			Target: neuronMap[oldConnection.Target],
			Weight: oldConnection.Weight,
		}
		if newConnection.Source == nil || newConnection.Target == nil {
			fmt.Printf("ERROR: nil source or target in new brain connection\n")
		}
		brain.Connections[i] = newConnection

		// Add this connection to the new neurons' connection slices
		neuronMap[oldConnection.Source].Connections = append(neuronMap[oldConnection.Source].Connections, newConnection)
	}

	//fmt.Printf("New brain created: %p\n", brain)
	return brain
}

type NeuronViewModel struct {
	ID    uint16	`json:"id"`
	Value float64	`json:"value"`
	Depth int		`json:"depth"`
}

type ConnectionViewModel struct {
	Source uint16	`json:"source"`
	Target uint16	`json:"target"`
	Weight float64	`json:"weight"`
}

type BrainViewModel struct {
	Neurons     []*NeuronViewModel		`json:"neurons"`
	Connections []*ConnectionViewModel	`json:"connections"`
}

func NewBrainViewModel(brain *Brain) *BrainViewModel {
	neurons := make([]*NeuronViewModel, 0, len(brain.InputNeurons)+len(brain.HiddenNeurons)+len(brain.OutputNeurons))
	connections := make([]*ConnectionViewModel, 0, len(brain.Connections))
	mapNeuronID := make(map[*Neuron]uint16)
	for i, neuron := range brain.InputNeurons {
		neurons = append(neurons, &NeuronViewModel{
			ID:    uint16(i),
			Value: neuron.Value,
			Depth: neuron.Depth,
		})
		mapNeuronID[neuron] = uint16(i)
	}

	for i, neuron := range brain.HiddenNeurons {
		neurons = append(neurons, &NeuronViewModel{
			ID:    uint16(i + len(brain.InputNeurons)),
			Value: neuron.Value,
			Depth: neuron.Depth,
		})
		mapNeuronID[neuron] = uint16(i + len(brain.InputNeurons))
	}

	for i, neuron := range brain.OutputNeurons {
		neurons = append(neurons, &NeuronViewModel{
			ID:    uint16(i + len(brain.InputNeurons) + len(brain.HiddenNeurons)),
			Value: neuron.Value,
			Depth: neuron.Depth,
		})
		mapNeuronID[neuron] = uint16(i + len(brain.InputNeurons) + len(brain.HiddenNeurons))
	}

	for _, connection := range brain.Connections {
		connections = append(connections, &ConnectionViewModel{
			Source: mapNeuronID[connection.Source],
			Target: mapNeuronID[connection.Target],
			Weight: connection.Weight,
		})
	}

	return &BrainViewModel{
		Neurons:     neurons,
		Connections: connections,
	}
}

func (b *Brain) connectionExists(source, target *Neuron) bool {
    for _, conn := range b.Connections {
        if conn.Source == source && conn.Target == target {
            return true
        }
    }
    return false
}