package mixerlib

import (
	"testing"
	"time"

	"github.com/ckaminer/jobcoin/clientlib"
	"github.com/stretchr/testify/assert"
)

// Begin processMixerUsers tests
func TestProcessMixerUsers_AddsUsersFromChannelToMixerUsers(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
		ReturnAddresses: []string{
			"1111aaaa",
			"2222bbbb",
			"3333cccc",
		},
	}

	MixerUsers = []MixerUser{}
	assert.Equal(t, 0, len(MixerUsers))

	// expectedErr := errors.New("GetAddressInfo failed")
	jobcoinMock := newJobcoinMock(clientlib.JobcoinAddressInfo{}, nil, nil)
	ml := &MixerLib{jobcoinMock}

	ticker := time.NewTicker(1 * time.Second)
	userChan := make(chan MixerUser, 1)
	userChan <- user

	ml.processMixerUsers(ticker, userChan, nil)

	assert.Equal(t, 1, len(MixerUsers))
	assert.Equal(t, user, MixerUsers[0])
}

func TestProcessMixerUsers_SendsUsersToHouseChannelIfFundsSentToHouse(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
		ReturnAddresses: []string{
			"1111aaaa",
			"2222bbbb",
			"3333cccc",
		},
	}
	MixerUsers = []MixerUser{user}

	// Balance greater than zero ensures user funds sent to house
	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "10",
	}
	jobcoinMock := newJobcoinMock(mockAddressInfo, nil, nil)
	ml := &MixerLib{jobcoinMock}

	ticker := time.NewTicker(1 * time.Second)
	houseChan := make(chan MixerUser, 1)

	ml.processMixerUsers(ticker, nil, houseChan)

	houseUser := <-houseChan

	assert.Equal(t, user, houseUser)
}

// Begin processHouseUsers tests
func TestProcessHouseUsers_AddsUsersFromChannelToHouseQueue(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
		ReturnAddresses: []string{
			"1111aaaa",
			"2222bbbb",
			"3333cccc",
		},
	}

	HouseQueue = []MixerUser{}
	assert.Equal(t, 0, len(HouseQueue))

	ml := &MixerLib{}

	ticker := time.NewTicker(1 * time.Second)
	houseChan := make(chan MixerUser, 1)
	houseChan <- user

	ml.processHouseUsers(ticker, houseChan)

	assert.Equal(t, 1, len(HouseQueue))
	assert.Equal(t, user, HouseQueue[0])
}

func TestProcessHouseUsers_RemoveUserFromHouseIfAllFundsReturned(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
		ReturnAddresses: []string{
			"1111aaaa",
			"2222bbbb",
			"3333cccc",
		},
	}
	HouseQueue = []MixerUser{user}

	// Amount sent to house from user less than deposit increment
	// All funds will be returned
	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "10",
		Transactions: []clientlib.JobcoinTx{
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "3",
			},
		},
	}
	jobcoinMock := newJobcoinMock(mockAddressInfo, nil, nil)
	ml := &MixerLib{jobcoinMock}

	ticker := time.NewTicker(1 * time.Second)

	ml.processHouseUsers(ticker, nil)

	assert.Equal(t, 0, len(HouseQueue))
}

func TestProcessHouseUsers_LeaveUserInHouseIfNotAllFundsReturned(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
		ReturnAddresses: []string{
			"1111aaaa",
			"2222bbbb",
			"3333cccc",
		},
	}
	HouseQueue = []MixerUser{user}

	// Amount sent to house from user more than deposit increment
	// Deposit increment will be returned
	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "500",
		Transactions: []clientlib.JobcoinTx{
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "100",
			},
		},
	}
	jobcoinMock := newJobcoinMock(mockAddressInfo, nil, nil)
	ml := &MixerLib{jobcoinMock}

	ticker := time.NewTicker(1 * time.Second)

	ml.processHouseUsers(ticker, nil)

	assert.Equal(t, 1, len(HouseQueue))
	assert.Equal(t, user, HouseQueue[0])
}
