package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ckaminer/jobcoin"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ckaminer/jobcoin/api"
	"github.com/ckaminer/jobcoin/clientlib"
	"github.com/ckaminer/jobcoin/mixerlib"
)

func main() {
	userChan := make(chan mixerlib.MixerUser)
	houseChan := make(chan mixerlib.MixerUser)

	r := mux.NewRouter()
	r.HandleFunc("/api/users", api.CreateNewUserHandler(userChan)).Methods("POST")

	userTicker := time.NewTicker(time.Second * 5)
	houseTicker := time.NewTicker(time.Second * 6)

	ml := &mixerlib.MixerLib{
		JobcoinClient: &clientlib.JobcoinLib{
			Client: &http.Client{},
		},
	}

	houseAddress, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	mixerlib.HouseAddress = houseAddress.String()
	fmt.Println("The house address has been set to: ", mixerlib.HouseAddress)

	go ml.PollForNewDeposits(userTicker, userChan, houseChan)
	go ml.PollForUserReturns(houseTicker, houseChan)

	log.Fatal(http.ListenAndServe(jobcoin.MixerPort, r))
}
