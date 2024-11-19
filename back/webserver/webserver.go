package webserver

import (
	"Prey_Predator_MAS/agents"
	"Prey_Predator_MAS/config"
	"Prey_Predator_MAS/simulation"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebServer struct {
	address       string
	port          string
	simulation    *simulation.Simulation
	selectedAgent *agents.Agent
	isPaused      bool
}

type SentData struct {
	Agents        []*agents.AgentViewModel `json:"agents"`
	TickCounter   uint64                   `json:"tickcounter"`
	PreyCount     int                      `json:"preycount"`
	PredatorCount int                      `json:"predatorcount"`
	ElapsedTime   int64                    `json:"elapsedtime"`
}

func NewWebServer(address string, port string) *WebServer {
	return &WebServer{
		address:    address,
		port:       port,
		simulation: simulation.NewSimulation(),
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                                // Allow any origin
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE") // Allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (wserver *WebServer) handleConnections(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	for {
		if wserver.isPaused {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		var agentsViewModels []*agents.AgentViewModel
		var selectedAgentVM *agents.AgentViewModel = nil

		if wserver.selectedAgent != nil {
			selectedAgentVM = agents.NewAgentViewModel(wserver.selectedAgent, true)
		}

		for _, agent := range wserver.simulation.Environment.Agents {
			if agent != wserver.selectedAgent && agent != nil {
				vm := agents.NewAgentViewModel(agent, false)
				agentsViewModels = append(agentsViewModels, vm)
			}
		}

		if selectedAgentVM != nil {
			agentsViewModels = append(agentsViewModels, selectedAgentVM)
		}

		wserver.simulation.Environment.LongPollIterationEnd()

		data := SentData{
			Agents:        agentsViewModels,
			TickCounter:   wserver.simulation.Environment.TickCounter,
			PreyCount:     wserver.simulation.Environment.PreyCount,
			PredatorCount: wserver.simulation.Environment.PredatorCount,
			ElapsedTime:   time.Since(wserver.simulation.Environment.StartTime).Milliseconds(),
		}

		err = ws.WriteJSON(data)

		if err != nil {
			return
		}
	}
}

func (wserver *WebServer) selectAgentInfo(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var selectRequest SelectRequest
	err := json.NewDecoder(r.Body).Decode(&selectRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, element := range wserver.simulation.Environment.Agents {
		if element.ID == uint32(selectRequest.AgentId) {
			wserver.selectedAgent = element
			return
		}
	}
	w.Write([]byte("Agent not found"))
}

func (wserver *WebServer) getConfig(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	val, err := json.Marshal(config.GetDefaultConfig())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(val)
}

func (wserver *WebServer) pause(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wserver.isPaused = true
}

func (wserver *WebServer) play(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wserver.isPaused = false
}

func (wserver *WebServer) Start() {

	go wserver.simulation.Start()
	// création du multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wserver.handleConnections)
	mux.Handle("/config", enableCORS(http.HandlerFunc(wserver.getConfig)))
	mux.Handle("/selectAgent", enableCORS(http.HandlerFunc(wserver.selectAgentInfo)))
	mux.Handle("/pause", enableCORS(http.HandlerFunc(wserver.pause)))
	mux.Handle("/play", enableCORS(http.HandlerFunc(wserver.play)))

	// création du serveur http
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", wserver.address, wserver.port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20}

	// lancement du serveur
	log.Println("Listening on", wserver.address)
	s.ListenAndServe()
}
