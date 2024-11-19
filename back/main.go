package main

import (
	"Prey_Predator_MAS/webserver"
	"math/rand"
	_ "net/http/pprof"
)

//import "net/http"
//import "github.com/pkg/profile"

func main() {
	rand.Seed(100000)
	//defer profile.Start(profile.ProfilePath(".")).Stop()
	// go func() {
	// 	http.ListenAndServe("localhost:6060", nil)
	// }()
	server := webserver.NewWebServer("localhost", "8080")
	server.Start()
}
