package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ckaminer/jobcoin/mixerlib"
	"github.com/google/uuid"
)

// ErrorPayload represents the error that will returned by the API.
type ErrorPayload struct {
	Message string `json:"error"`
}

// CreateNewUserHandler returns a HandlerFunc to handle the creation of users.
// It accepts a channel to be used in the resulting HanderFunc
// HandlerFunc will validate inputs before sending users into the provided userChannel.
func CreateNewUserHandler(userChan chan mixerlib.MixerUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user mixerlib.MixerUser
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Println("NewUserHandler error: ", err.Error())
			respondWithJSON(w, http.StatusBadRequest, ErrorPayload{"Invalid request body"})
			return
		}
		defer r.Body.Close()

		if address, valid := mixerlib.ValidUserAddresses(user.ReturnAddresses); !valid {
			respondWithJSON(w, http.StatusConflict, ErrorPayload{
				fmt.Sprintf("Return address %s is already in use", address),
			})
			return
		}

		depositAddress, err := uuid.NewUUID()
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, ErrorPayload{"Failed to create user"})
			log.Fatal(err)
		}
		user.DepositAddress = depositAddress.String()

		userChan <- user

		respondWithJSON(w, http.StatusCreated, user)
	}
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}
