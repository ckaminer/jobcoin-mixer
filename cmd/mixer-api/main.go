package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ckaminer/jobcoin"

	"github.com/ckaminer/jobcoin/api"
	"github.com/ckaminer/jobcoin/clientlib"
	"github.com/ckaminer/jobcoin/mixerlib"
)

func main() {
	http.HandleFunc("/api/users", api.NewUserHandler)

	api.UserChannel = make(chan mixerlib.MixerUser)
	api.HouseChannel = make(chan mixerlib.MixerUser)

	userTicker := time.NewTicker(time.Second * 5)
	houseTicker := time.NewTicker(time.Second * 6)

	ml := &mixerlib.MixerLib{
		JobcoinClient: &clientlib.JobcoinLib{
			Client: &http.Client{},
		},
	}
	go ml.PollForNewDeposits(userTicker, api.UserChannel, api.HouseChannel)
	go ml.PollForUserReturns(houseTicker, api.HouseChannel)

	log.Fatal(http.ListenAndServe(jobcoin.MixerPort, nil))
}
