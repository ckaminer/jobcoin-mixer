package clientlib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ckaminer/jobcoin"
)

// HTTPClient is an interface representing functionality of an http client
// required for this package.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// JobcoinTx represents the transactions that are returned from and sent to the jobcoin rest api
type JobcoinTx struct {
	Timestamp   string `json:"timestamp"`
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	Amount      string `json:"amount"`
}

// JobcoinAddressInfo represents info for a Jobcoin address
type JobcoinAddressInfo struct {
	Balance      string      `json:"balance"`
	Transactions []JobcoinTx `json:"transactions"`
}

// JobcoinClient is an interface respresenting functionality needed to
// interact with the Jobcoin API.
type JobcoinClient interface {
	GetAddressInfo(address string) (JobcoinAddressInfo, error)
	SendJobcoin(fromAddress, toAddress, amount string) error
}

// JobcoinLib is an implementation of the JobcoinClient interface. It requires
// an HTTPClient to make network calls.
type JobcoinLib struct {
	Client HTTPClient
}

// GetAddressInfo should return address info for given address
func (jl *JobcoinLib) GetAddressInfo(address string) (JobcoinAddressInfo, error) {
	url := fmt.Sprintf("%s/%s", jobcoin.AddressesEndpoint, address)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return JobcoinAddressInfo{}, err
	}

	res, err := jl.Client.Do(req)
	if err != nil {
		log.Println(err)
		return JobcoinAddressInfo{}, err
	}
	defer res.Body.Close()

	var addrInfo JobcoinAddressInfo
	err = json.NewDecoder(res.Body).Decode(&addrInfo)
	if err != nil {
		log.Println(err)
		return JobcoinAddressInfo{}, err
	}

	return addrInfo, nil
}

// SendJobcoin creates a transaction sending the specified amount between the given addresses
func (jl *JobcoinLib) SendJobcoin(fromAddress, toAddress, amount string) error {
	reqBody, err := json.Marshal(JobcoinTx{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	})
	if err != nil {
		log.Println(err)
		return err
	}

	req, err := http.NewRequest("POST", jobcoin.TransactionEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := jl.Client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		apiErr := struct {
			Error interface{} `json:"error"`
		}{}
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		if err != nil {
			log.Println(err)
			return err
		}
		errMessage := fmt.Sprintf("Failed to create transaction due to: %s", apiErr.Error)
		log.Println(errMessage)
		return errors.New(errMessage)
	}

	return nil
}
