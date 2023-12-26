package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	playerRegistration "github.com/p-society/gCSB/connector/controllers"
	generaldb "github.com/p-society/gCSB/connector/db"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/register/player", playerRegistration.PlayerRegistrationController).Methods("POST")
	r.HandleFunc("/verify/player", playerRegistration.VerifyPlayerController).Methods("POST")
	r.HandleFunc("/captain/fetch", playerRegistration.CaptainFetchController).Methods("POST")
	r.HandleFunc("/captain/upload-roster", playerRegistration.CaptainPushController).Methods("POST")
	
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		generaldb.Init()
		err := http.ListenAndServe(":45234", r)

		if err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Connector Live at Port :45234")
	wg.Wait()
}
