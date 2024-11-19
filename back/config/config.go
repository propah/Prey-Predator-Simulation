package config

const WIDTH = 1024
const HEIGHT = 1024
const NUM_AGENTS = 1000
const MAX_PREY = 2000
const MAX_PREDATOR = 600
const CELL_SIZE = 8
const APPROXIMATE_MAX_AGENT_PER_CELL = 400 //NUM_AGENTS / ((HEIGHT / CELL_SIZE) * (WIDTH / CELL_SIZE)) * 4
const AGENT_RADIUS = 1
const RAY_NUMBER = 24
const PREDATOR_RAY_LENGTH = 80
const PREY_RAY_LENGTH = 40
const PREDATOR_RAY_ANGLE_DEG = 30
const PREY_RAY_ANGLE_DEG = 250

const FRONT_SCALE_FACTOR = 5

const MAX_ENERGY = 550
const MAX_SPEED = 2

const PREY_LIFE_POINTS = 1
const PREDATOR_LIFE_POINTS = 5

const PREY_ATTACK_DAMAGE = 1
const PREDATOR_ATTACK_DAMAGE = 3

const PREDATOR_ENERGY_GAIN = MAX_ENERGY
const PREY_ENERGY_GAIN = 10

const PREY_REPRODUCTION_GAIN = 2
const MAX_REPRODUCTION_PREY = 200
const MAX_REPRODUCTION_PREDATOR = 300
const PREDATOR_REPRODUCTION_GAIN = MAX_REPRODUCTION_PREDATOR

const ENERGY_LOSS_MULTIPLIER_SPEED = 3

// NEURONS
const INPUT_NEURON_NUMBER = RAY_NUMBER
const OUTPUT_NEURON_NUMBER = 2
const MAX_NEURON_NUMBER = 50

const NO_MUTATION = 20
const WEIGHT_MUTATION_RATE = 20
const BIAS_MUTATION_RATE = 10
const NEW_CONNECTION_RATE = 60
const DEL_CONNECTION_RATE = 5
const NEW_NEURON_RATE = 5
const DEL_NEURON_RATE = 1
const START_MUTATION_NUMBER = 0
const WEIGHT_MUTATION_STAND_DEV = 5
const BIAS_MUTATION_STAND_DEV = 0.1

type Config struct {
	Width               int `json:"width"`
	Height              int `json:"height"`
	NumAgents           int `json:"numAgents"`
	CellSize            int `json:"cellSize"`
	AgentRadius         int `json:"agentRadius"`
	RayNumber           int `json:"rayNumber"`
	PredatorRayLength   int `json:"predatorRayLength"`
	PreyRayLength       int `json:"preyRayLength"`
	PredatorRayAngleDeg int `json:"predatorRayAngleDeg"`
	PreyRayAngleDeg     int `json:"preyRayAngleDeg"`
	InputNeuronNumber   int `json:"inputNeuronNumber"`
	OutputNeuronNumber  int `json:"outputNeuronNumber"`
	PreyLifePoints      int `json:"preyLifePoints"`
	PredatorLifePoints  int `json:"predatorLifePoints"`

	PreyAttackDamage          int `json:"preyAttackDamage"`
	PredatorAttackDamage      int `json:"predatorAttackDamage"`
	PreyEnergyGain            int `json:"preyEnergyGain"`
	PredatorEnergyGain        int `json:"predatorEnergyGain"`
	PreyReproductionGain      int `json:"preyReproductionGain"`
	PredatorReproductionGain  int `json:"predatorReproductionGain"`
	EnergyLossMultiplierSpeed int `json:"energyLossMultiplierSpeed"`
	MaxSpeed                  int `json:"maxSpeed"`

	ScaleFactor int `json:"scaleFactor"`
	MaxPrey     int `json:"maxPrey"`
	MaxPredator int `json:"maxPredator"`
	MaxEnergy   int `json:"maxEnergy"`
}

type MutationRate struct {
	NoMutation         int `json:"noMutation"`
	WeightMutationRate int `json:"weightMutationRate"`
	BiasMutationRate   int `json:"biasMutationRate"`
	NewConnectionRate  int `json:"newConnectionRate"`
	DelConnectionRate  int `json:"delConnectionRate"`
	NewNeuronRate      int `json:"newNeuronRate"`
	DelNeuronRate      int `json:"delNeuronRate"`
}

func GetDefaultConfig() Config {
	return Config{
		Width:                     WIDTH,
		Height:                    HEIGHT,
		NumAgents:                 NUM_AGENTS,
		CellSize:                  CELL_SIZE,
		AgentRadius:               AGENT_RADIUS,
		RayNumber:                 RAY_NUMBER,
		PredatorRayLength:         PREDATOR_RAY_LENGTH,
		PreyRayLength:             PREY_RAY_LENGTH,
		PredatorRayAngleDeg:       PREDATOR_RAY_ANGLE_DEG,
		PreyRayAngleDeg:           PREY_RAY_ANGLE_DEG,
		InputNeuronNumber:         INPUT_NEURON_NUMBER,
		OutputNeuronNumber:        OUTPUT_NEURON_NUMBER,
		PreyLifePoints:            PREY_LIFE_POINTS,
		PredatorLifePoints:        PREDATOR_LIFE_POINTS,
		PreyAttackDamage:          PREY_ATTACK_DAMAGE,
		PredatorAttackDamage:      PREDATOR_ATTACK_DAMAGE,
		PreyEnergyGain:            PREY_ENERGY_GAIN,
		PredatorEnergyGain:        PREDATOR_ENERGY_GAIN,
		PreyReproductionGain:      PREY_REPRODUCTION_GAIN,
		PredatorReproductionGain:  PREDATOR_REPRODUCTION_GAIN,
		EnergyLossMultiplierSpeed: ENERGY_LOSS_MULTIPLIER_SPEED,
		MaxSpeed:                  MAX_SPEED,
		ScaleFactor:               FRONT_SCALE_FACTOR,
		MaxPrey:                   MAX_PREY,
		MaxPredator:               MAX_PREDATOR,
		MaxEnergy:                 MAX_ENERGY,
	}
}

func GetDefaultMutationRate() MutationRate {
	return MutationRate{
		NoMutation:         NO_MUTATION,
		WeightMutationRate: WEIGHT_MUTATION_RATE,
		BiasMutationRate:   BIAS_MUTATION_RATE,
		NewConnectionRate:  NEW_CONNECTION_RATE,
		DelConnectionRate:  DEL_CONNECTION_RATE,
		NewNeuronRate:      NEW_NEURON_RATE,
		DelNeuronRate:      DEL_NEURON_RATE,
	}
}
