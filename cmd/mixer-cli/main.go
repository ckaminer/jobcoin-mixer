package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ckaminer/jobcoin/api"

	"github.com/ckaminer/jobcoin"
	"github.com/ckaminer/jobcoin/mixerlib"
	"github.com/google/uuid"
)

func inputDepositAddresses() []string {

	var svar string
	flag.StringVar(&svar, "addresses", "", "comma-separated list of new, unused Jobcoin addresses")

	flag.Parse()

	trimmed := strings.TrimSpace(svar)
	if trimmed == "" {
		instruction := `
Welcome to the Jobcoin mixer!
Please enter a comma-separated list of new, unused Jobcoin addresses
where your mixed Jobcoins will be sent. Example:

	./bin/mixer --addresses=bravo,tango,delta
`
		fmt.Println(instruction)
		os.Exit(-1)
	}

	return strings.Split(strings.ToLower(trimmed), ",")
}

func createMixerUser(user mixerlib.MixerUser) error {
	reqBody, err := json.Marshal(user)
	if err != nil {
		log.Println("Error creating request body: ", err)
		return err
	}

	req, _ := http.NewRequest("POST", jobcoin.MixerUserEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		log.Println("Error creating request: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Println("Request error: ", err)
		return err
	}
	if res.StatusCode != http.StatusCreated {
		var apiErr api.ErrorPayload
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		if err != nil {
			log.Println("Error decoding api response: ", err)
			return err
		}
		return errors.New(apiErr.Message)
	}

	return nil
}

func main() {
	addresses := inputDepositAddresses()
	depositAddress, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}

	user := mixerlib.MixerUser{
		DepositAddress:  depositAddress.String(),
		ReturnAddresses: addresses,
	}

	err = createMixerUser(user)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf(`
You may now send Jobcoins to address %s.

They will be mixed into %s and sent to your destination addresses.`, depositAddress, addresses)
}
