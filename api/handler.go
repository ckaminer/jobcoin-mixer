package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ckaminer/jobcoin/mixerlib"
)

// UserChannel is used to introduce created users into the Mixer poll for new users
var UserChannel chan mixerlib.MixerUser

// HouseChannel is used to introduce users into the HouseQueue poll
var HouseChannel chan mixerlib.MixerUser

// ErrorPayload represents the error that will returned by the API.
type ErrorPayload struct {
	Message string `json:"error"`
}

// NewUserHandler handles the creation of users.
// It will validate inputs before sending users into the UserChannel.
func NewUserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var user mixerlib.MixerUser
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Println("NewUserHandler error: ", err.Error())
			respondWithJSON(w, http.StatusBadRequest, ErrorPayload{"Invalid request body"})
			return
		}

		if address, valid := mixerlib.ValidUserAddresses(user.ReturnAddresses); !valid {
			respondWithJSON(w, http.StatusConflict, ErrorPayload{
				fmt.Sprintf("Return address %s is already in use", address),
			})
			return
		}

		UserChannel <- user

		respondWithJSON(w, http.StatusCreated, user)
		defer r.Body.Close()
	default:
		respondWithJSON(w, http.StatusNotFound, nil)
	}
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}
