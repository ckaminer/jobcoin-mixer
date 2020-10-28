package mixerlib

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/ckaminer/jobcoin/clientlib"
)

// MixerUsers is a collection of user information.
// When users use the Jobcoin mixer their information will be stored here.
var MixerUsers = []MixerUser{}

// HouseQueue is a collection of users whose money has been moved from
// their deposit address to the house address. It is a subset of MixerUsers.
var HouseQueue = []MixerUser{}

// HouseAddress is the address used for the house account. This will
// be the Jobcoin repository for user submitted coins though no user
// transactions should send Jobcoins directly to the house address.
const HouseAddress = "1234abcd5678efgh"

// DistributionIncrement represents the amount of Jobcoin that will be returned
// back to users during each round of returns.
const DistributionIncrement = 5.0

// MixerUser organizes addresses and transactions for a client of the Jobcoin Mixer
type MixerUser struct {
	DepositAddress  string   `json:"depositAddress"`
	ReturnAddresses []string `json:"returnAddresses"`
}

// MixerClient is an interface respresenting functionality needed to
// interact with the Jobcoin Mixer.
type MixerClient interface {
	PollForNewDeposits(ticker *time.Ticker, userChan, houseChan chan MixerUser)
	PollForUserReturns(ticker *time.Ticker, houseChan chan MixerUser)
}

// MixerLib is an implementation of the MixerClient interface. It requires
// a JobcoinClient to interact with the Jobcoin API.
type MixerLib struct {
	JobcoinClient clientlib.JobcoinClient
}

func (ml *MixerLib) transferDepositToHouse(user MixerUser) (bool, error) {
	sentToHouse := false

	info, err := ml.JobcoinClient.GetAddressInfo(user.DepositAddress)
	if err != nil {
		return sentToHouse, err
	}

	balance, _ := strconv.ParseFloat(info.Balance, 64)
	if balance > 0 {
		sentToHouse = true
		err = ml.JobcoinClient.SendJobcoin(user.DepositAddress, HouseAddress, info.Balance)
		if err != nil {
			return false, err
		}
	}

	return sentToHouse, nil
}

func (ml *MixerLib) returnFundsToUser(user MixerUser) (bool, error) {
	sendingEntireBalance := true

	houseBalance, err := ml.calculateHouseBalanceForUser(user)
	if err != nil {
		return false, err
	}

	if houseBalance > 0 {
		distAmount := houseBalance
		if houseBalance > DistributionIncrement {
			distAmount = DistributionIncrement
			sendingEntireBalance = false
		}

		returnAmounts := ml.assignReturnAmounts(user.ReturnAddresses, distAmount)

		for address, amount := range returnAmounts {
			err := ml.JobcoinClient.SendJobcoin(HouseAddress, address, amount)
			if err != nil {
				sendingEntireBalance = false
				continue
			}
		}
	}
	return sendingEntireBalance, nil
}

func (ml *MixerLib) calculateHouseBalanceForUser(user MixerUser) (float64, error) {
	houseInfo, err := ml.JobcoinClient.GetAddressInfo(HouseAddress)
	if err != nil {
		return 0, err
	}

	var userHouseDepositTotal float64
	var returnedToUserTotal float64

	for _, tx := range houseInfo.Transactions {
		if tx.FromAddress == user.DepositAddress && tx.ToAddress == HouseAddress {
			amount, _ := strconv.ParseFloat(tx.Amount, 64)
			userHouseDepositTotal = userHouseDepositTotal + amount
		}

		sentToUserAddress := containsElement(user.ReturnAddresses, tx.ToAddress)
		if tx.FromAddress == HouseAddress && sentToUserAddress {
			amount, _ := strconv.ParseFloat(tx.Amount, 64)
			returnedToUserTotal = returnedToUserTotal + amount
		}
	}

	return userHouseDepositTotal - returnedToUserTotal, nil
}

func (ml *MixerLib) assignReturnAmounts(addresses []string, distAmount float64) map[string]string {
	rand.Seed(time.Now().UnixNano())
	returnAmounts := make(map[string]string)

	if len(addresses) > 0 {
		// Randomly assign values to all but last address
		for i := 0; i < len(addresses)-1; i++ {
			returnAddress := addresses[i]
			if distAmount < 0.0001 {
				amount := fmt.Sprintf("%g", distAmount)
				returnAmounts[returnAddress] = amount
				break
			}
			randAmount := distAmount * 0.9 * rand.Float64()
			if randAmount > 0 {
				amount := fmt.Sprintf("%g", randAmount)
				returnAmounts[returnAddress] = amount
			}

			distAmount = distAmount - randAmount
		}

		// Assign reminder to last address
		returnAmounts[addresses[len(addresses)-1]] = fmt.Sprintf("%f", distAmount)
	}

	return returnAmounts
}

// ValidUserAddresses returns a thing
func ValidUserAddresses(addresses []string) (string, bool) {
	allReturnAddresses := []string{}
	for _, user := range MixerUsers {
		allReturnAddresses = append(allReturnAddresses, user.ReturnAddresses...)
	}

	for _, address := range addresses {
		if containsElement(allReturnAddresses, address) {
			return address, false
		}
	}

	return "", true
}

func containsElement(collection []string, item string) bool {
	for _, elem := range collection {
		if elem == item {
			return true
		}
	}
	return false
}

func addOrReplaceUserInCollection(collection []MixerUser, user MixerUser) []MixerUser {
	existingIdx := -1
	for i, c := range collection {
		if c.DepositAddress == user.DepositAddress {
			existingIdx = i
		}
	}

	if existingIdx != -1 {
		collection = append(collection[:existingIdx], collection[existingIdx+1:]...)
	}
	return append(collection, user)
}
