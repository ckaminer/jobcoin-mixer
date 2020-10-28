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
	"github.com/ckaminer/jobcoin/clientlib"

	"github.com/ckaminer/jobcoin"
	"github.com/ckaminer/jobcoin/mixerlib"
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

	./bin/mixer-cli --addresses=bravo,tango,delta
`
		fmt.Println(instruction)
		os.Exit(-1)
	}

	return strings.Split(strings.ToLower(trimmed), ",")
}

func createMixerUser(client clientlib.HTTPClient, returnAddresses []string) (mixerlib.MixerUser, error) {
	reqBody, err := json.Marshal(mixerlib.MixerUser{ReturnAddresses: returnAddresses})
	if err != nil {
		log.Println("Error creating request body: ", err)
		return mixerlib.MixerUser{}, err
	}

	req, _ := http.NewRequest("POST", jobcoin.MixerUserEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		log.Println("Error creating request: ", err)
		return mixerlib.MixerUser{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Println("Request error: ", err)
		return mixerlib.MixerUser{}, err
	}
	if res.StatusCode != http.StatusCreated {
		var apiErr api.ErrorPayload
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		if err != nil {
			log.Println("Error decoding api response: ", err)
			return mixerlib.MixerUser{}, err
		}
		defer res.Body.Close()
		return mixerlib.MixerUser{}, errors.New(apiErr.Message)
	}

	var createdUser mixerlib.MixerUser
	err = json.NewDecoder(res.Body).Decode(&createdUser)
	if err != nil {
		log.Println("Error decoding api response: ", err)
		return mixerlib.MixerUser{}, err
	}
	defer res.Body.Close()

	return createdUser, nil
}

func main() {
	addresses := inputDepositAddresses()
	client := &http.Client{}

	createdUser, err := createMixerUser(client, addresses)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`
You may now send Jobcoins to address %s.

They will be mixed into %s and sent to your destination addresses.`, createdUser.DepositAddress, createdUser.ReturnAddresses)
}
