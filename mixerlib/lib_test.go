package mixerlib

import (
	"errors"
	"strconv"
	"testing"

	"github.com/ckaminer/jobcoin/clientlib"
	"github.com/stretchr/testify/assert"
)

// Begin transferDepositToHouse tests
func TestTransferDepositToHouse_ReturnsTrueIfBalanceSentToHouse(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
		ReturnAddresses: []string{
			"1111aaaa",
			"2222bbbb",
			"3333cccc",
		},
	}

	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "0.54",
	}
	jobcoinMock := newJobcoinMock(mockAddressInfo, nil, nil)

	ml := &MixerLib{jobcoinMock}

	sentToHouse, err := ml.transferDepositToHouse(user)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.True(t, sentToHouse)
}

func TestTransferDepositToHouse_DoesNotCreateTransactionIfBalanceZero(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
		ReturnAddresses: []string{
			"1111aaaa",
			"2222bbbb",
			"3333cccc",
		},
	}

	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "0",
	}
	// Make SendJobcoin return error since this test should not be calling that function.
	sendErr := errors.New("SendJobcoin failed")
	jobcoinMock := newJobcoinMock(mockAddressInfo, nil, sendErr)

	ml := &MixerLib{jobcoinMock}

	sentToHouse, err := ml.transferDepositToHouse(user)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.False(t, sentToHouse)
}

func TestTransferDepositToHouse_ReturnsErrorIfInfoRetrievalFails(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
	}

	expectedErr := errors.New("GetAddressInfo failed")
	jobcoinMock := newJobcoinMock(clientlib.JobcoinAddressInfo{}, expectedErr, nil)

	ml := &MixerLib{jobcoinMock}

	sentToHouse, err := ml.transferDepositToHouse(user)
	if err == nil {
		t.Errorf("Expected error to be returned but it was not.")
	}

	assert.Equal(t, expectedErr, err)
	assert.False(t, sentToHouse)
}

func TestTransferDepositToHouse_ReturnsErrorIfSendJobcoinFails(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1234abcd",
	}

	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "1",
	}
	expectedErr := errors.New("SendJobcoin failed")
	jobcoinMock := newJobcoinMock(mockAddressInfo, nil, expectedErr)

	ml := &MixerLib{jobcoinMock}

	sentToHouse, err := ml.transferDepositToHouse(user)
	if err == nil {
		t.Errorf("Expected error to be returned but it was not.")
	}

	assert.Equal(t, expectedErr, err)
	assert.False(t, sentToHouse)

}

// Begin returnFundsToUser tests
func TestReturnFundsToUser_ReturnsTrueIfFullBalanceReturned(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1111aaaa",
		ReturnAddresses: []string{
			"2222bbbb",
			"3333cccc",
		},
	}

	// Transactions from user's deposit address to house total up to less than return increment
	// No funds have returned back to user yet
	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "1000.045",
		Transactions: []clientlib.JobcoinTx{
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "3.045",
			},
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "1.324",
			},
		},
	}

	jc := newJobcoinMock(mockAddressInfo, nil, nil)
	ml := MixerLib{jc}

	emptyBalance, err := ml.returnFundsToUser(user)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.True(t, emptyBalance)
}

func TestReturnFundsToUser_ReturnsFalseIfFullBalanceNotReturned(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1111aaaa",
		ReturnAddresses: []string{
			"2222bbbb",
			"3333cccc",
		},
	}

	// Transactions from user's deposit address to house total up to more than return increment
	// Some funds have returned back to user yet
	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "1000.045",
		Transactions: []clientlib.JobcoinTx{
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "3.045",
			},
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "14.324",
			},
			{
				FromAddress: HouseAddress,
				ToAddress:   user.ReturnAddresses[0],
				Amount:      "2.8623",
			},
		},
	}

	jc := newJobcoinMock(mockAddressInfo, nil, nil)
	ml := MixerLib{jc}

	emptyBalance, err := ml.returnFundsToUser(user)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.False(t, emptyBalance)
}

func TestReturnFundsToUser_ErrorsIfUnableToRetrieveUserInfo(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1111aaaa",
		ReturnAddresses: []string{
			"2222bbbb",
			"3333cccc",
		},
	}

	userError := errors.New("Unable to retrieve user info")

	jc := newJobcoinMock(clientlib.JobcoinAddressInfo{}, userError, nil)
	ml := MixerLib{jc}

	_, err := ml.returnFundsToUser(user)
	if err == nil {
		t.Errorf("Expectd error but did not receive one.")
	}
}

func TestReturnFundsToUser_ReturnsFalseIfFailsToCreateTransaction(t *testing.T) {
	user := MixerUser{
		DepositAddress: "1111aaaa",
		ReturnAddresses: []string{
			"2222bbbb",
			"3333cccc",
		},
	}

	// Transactions from user's deposit address to house total up to less than return increment
	// No funds have returned back to user yet
	// With no transaction failures, this would typically cause true to be returned
	mockAddressInfo := clientlib.JobcoinAddressInfo{
		Balance: "1000.045",
		Transactions: []clientlib.JobcoinTx{
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "3.045",
			},
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "1.324",
			},
		},
	}

	sendError := errors.New("Unable to send Jobcoin")

	jc := newJobcoinMock(mockAddressInfo, nil, sendError)
	ml := MixerLib{jc}

	emptyBalance, err := ml.returnFundsToUser(user)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.False(t, emptyBalance)
}

// Begin calculateHouseBalanceForUser tests
func TestCalculateHouseBalanceForUser_ReturnsDiffOfHouseDepositAndReturns(t *testing.T) {
	user := MixerUser{
		DepositAddress: "user-deposit-address",
		ReturnAddresses: []string{
			"return-address-1",
			"return-address-2",
			"return-address-3",
		},
	}

	houseInfo := clientlib.JobcoinAddressInfo{
		Balance: "168.37",
		Transactions: []clientlib.JobcoinTx{
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "10.79485481",
			},
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "5.23684368",
			},
			{
				FromAddress: HouseAddress,
				ToAddress:   user.ReturnAddresses[0],
				Amount:      "1.96342871",
			},
			{
				FromAddress: HouseAddress,
				ToAddress:   user.ReturnAddresses[1],
				Amount:      "3.56385431",
			},
			{
				FromAddress: HouseAddress,
				ToAddress:   user.ReturnAddresses[2],
				Amount:      "2.99763821",
			},
		},
	}

	jobcoinMock := newJobcoinMock(houseInfo, nil, nil)
	ml := &MixerLib{jobcoinMock}

	expectedBalance := 7.50677726
	actualBalance, err := ml.calculateHouseBalanceForUser(user)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.Equal(t, expectedBalance, actualBalance)
}

func TestCalculateHouseBalanceForUser_ReturnsDiff_Zero(t *testing.T) {
	user := MixerUser{
		DepositAddress: "user-deposit-address",
		ReturnAddresses: []string{
			"return-address-1",
			"return-address-2",
			"return-address-3",
		},
	}

	houseInfo := clientlib.JobcoinAddressInfo{
		Balance: "168.37",
		Transactions: []clientlib.JobcoinTx{
			{
				FromAddress: user.DepositAddress,
				ToAddress:   HouseAddress,
				Amount:      "9.97",
			},
			{
				FromAddress: HouseAddress,
				ToAddress:   user.ReturnAddresses[0],
				Amount:      "3.324",
			},
			{
				FromAddress: HouseAddress,
				ToAddress:   user.ReturnAddresses[1],
				Amount:      "3.323",
			},
			{
				FromAddress: HouseAddress,
				ToAddress:   user.ReturnAddresses[2],
				Amount:      "3.323",
			},
		},
	}

	jobcoinMock := newJobcoinMock(houseInfo, nil, nil)
	ml := &MixerLib{jobcoinMock}

	expectedBalance := 0.0
	actualBalance, err := ml.calculateHouseBalanceForUser(user)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.Equal(t, expectedBalance, actualBalance)
}

func TestCalculateHouseBalanceForUser_ReturnsErrorIfUnableToRetrieveInfo(t *testing.T) {
	user := MixerUser{
		DepositAddress: "user-deposit-address",
		ReturnAddresses: []string{
			"return-address-1",
			"return-address-2",
			"return-address-3",
		},
	}

	expectedErr := errors.New("Failed to get address info")
	jobcoinMock := newJobcoinMock(clientlib.JobcoinAddressInfo{}, expectedErr, nil)
	ml := &MixerLib{jobcoinMock}

	_, err := ml.calculateHouseBalanceForUser(user)
	if err == nil {
		t.Errorf("Expected error to be returned but it was not.")
	}

	assert.Equal(t, expectedErr, err)
}

// Being assignReturnAmounts tests
func TestAssignReturnAmounts_ReturnsMapOfAddressToRandomReturnAmounts(t *testing.T) {
	addresses := []string{
		"1111aaaa",
		"2222bbbb",
		"3333cccc",
	}
	distAmount := 5.0

	ml := &MixerLib{}
	returnAmounts := ml.assignReturnAmounts(addresses, distAmount)

	for _, returnAmount := range returnAmounts {
		amount, _ := strconv.ParseFloat(returnAmount, 64)
		assert.GreaterOrEqual(t, amount, 0.01)
		assert.LessOrEqual(t, amount, distAmount)
	}
}

func TestAssignReturnAmounts_DistributesEntireAmountIfOneAddress(t *testing.T) {
	addresses := []string{
		"1111aaaa",
	}
	distAmount := 5.0

	expectedReturns := map[string]string{
		"1111aaaa": "5.000000",
	}

	ml := &MixerLib{}
	returnAmounts := ml.assignReturnAmounts(addresses, distAmount)

	assert.Equal(t, expectedReturns, returnAmounts)
}

func TestAssignReturnAmounts_ReturnsEmptyMapIfNoAddress(t *testing.T) {
	addresses := []string{}
	distAmount := 5.0

	expectedReturns := map[string]string{}

	ml := &MixerLib{}
	returnAmounts := ml.assignReturnAmounts(addresses, distAmount)

	assert.Equal(t, expectedReturns, returnAmounts)
}

// Begin ValidUserAddresses tests
func TestValidUserAddresses_ReturnsTrueIfAllUniqueAddresses(t *testing.T) {
	userOne := MixerUser{
		ReturnAddresses: []string{
			"return-one",
			"return-two",
		},
	}
	userTwo := MixerUser{
		ReturnAddresses: []string{
			"return-three",
			"return-four",
			"return-five",
		},
	}

	MixerUsers = []MixerUser{userOne, userTwo}

	newAddresses := []string{"return-six", "return-seven"}

	_, valid := ValidUserAddresses(newAddresses)

	assert.True(t, valid)
}

func TestValidUserAddresses_ReturnsFalseAndCulpritIfNotUnique(t *testing.T) {
	userOne := MixerUser{
		ReturnAddresses: []string{
			"return-one",
			"return-two",
		},
	}
	userTwo := MixerUser{
		ReturnAddresses: []string{
			"return-three",
			"return-four",
			"return-five",
		},
	}

	MixerUsers = []MixerUser{userOne, userTwo}

	newAddresses := []string{"return-six", "return-five"}

	badAddress, valid := ValidUserAddresses(newAddresses)

	assert.False(t, valid)
	assert.Equal(t, "return-five", badAddress)
}

// Begin containsElement tests
func TestContainsElement_ReturnsTrueIfElementInSlice(t *testing.T) {
	collection := []string{"1111aaaa", "2222bbbb", "3333cccc"}
	containsItem := containsElement(collection, "1111aaaa")
	assert.True(t, containsItem)
}

func TestContainsElement_ReturnsFalseIfElementNotInSlice(t *testing.T) {
	collection := []string{"1111aaaa", "2222bbbb", "3333cccc"}
	containsItem := containsElement(collection, "4444dddd")
	assert.False(t, containsItem)
}

// Begin addOrReplaceUserInCollection tests
func TestAddOrReplaceUserInCollection_BumpsUserToEndIfAlreadyInCollection(t *testing.T) {
	collection := []MixerUser{
		{DepositAddress: "aaa"},
		{DepositAddress: "bbb"},
		{DepositAddress: "ccc"},
	}
	user := MixerUser{DepositAddress: "bbb"}

	expectedCollection := []MixerUser{
		{DepositAddress: "aaa"},
		{DepositAddress: "ccc"},
		{DepositAddress: "bbb"},
	}
	actualCollection := addOrReplaceUserInCollection(collection, user)

	assert.Equal(t, expectedCollection, actualCollection)
}

func TestAddOrReplaceUserInCollection_AddsUserToEndOfQueueIfNotInCollection(t *testing.T) {
	collection := []MixerUser{
		{DepositAddress: "aaa"},
		{DepositAddress: "bbb"},
	}
	user := MixerUser{DepositAddress: "ccc"}

	expectedCollection := []MixerUser{
		{DepositAddress: "aaa"},
		{DepositAddress: "bbb"},
		{DepositAddress: "ccc"},
	}
	actualCollection := addOrReplaceUserInCollection(collection, user)

	assert.Equal(t, expectedCollection, actualCollection)
}
