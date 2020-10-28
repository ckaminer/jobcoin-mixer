package clientlib

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Begin GetAddressInfo tests
func TestGetAddressInfo_ReturnsAddressInfoForGivenAddress(t *testing.T) {
	mockResponseBody := []byte(`{
		"balance": "10.53",
		"transactions": [
			{
				"timestamp": "2020-10-23T14:05:01.199Z",
				"toAddress": "01234abcde",
				"amount": "20.97"
			},
			{
				"timestamp": "2020-10-24T09:29:51.320Z",
				"fromAddress": "01234abcde",
				"toAddress": "98765zyxwt",
				"amount": "10.46"
			}
		]
	}`)
	client := newClientMock(http.StatusOK, mockResponseBody, nil)
	jl := &JobcoinLib{client}

	expectedAddrInfo := JobcoinAddressInfo{
		Balance: "10.53",
		Transactions: []JobcoinTx{
			{
				Timestamp:   "2020-10-23T14:05:01.199Z",
				FromAddress: "",
				ToAddress:   "01234abcde",
				Amount:      "20.97",
			},
			{
				Timestamp:   "2020-10-24T09:29:51.320Z",
				FromAddress: "01234abcde",
				ToAddress:   "98765zyxwt",
				Amount:      "10.46",
			},
		},
	}

	actualAddrInfo, err := jl.GetAddressInfo("01234abcde")
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.Equal(t, expectedAddrInfo, actualAddrInfo)
}

func TestGetAddressInfo_ReturnsErrorIfClientRequestFails(t *testing.T) {
	expectedErr := errors.New("request failed")
	client := newClientMock(0, nil, expectedErr)
	jl := &JobcoinLib{client}

	_, err := jl.GetAddressInfo("01234abcde")
	if err == nil {
		t.Errorf("Expected error to be returned but it was not.")
	}

	assert.Equal(t, expectedErr, err)
}

func TestGetAddressInfo_ReturnsErrorIfJsonFailsToDecode(t *testing.T) {
	mockResponseBody := []byte(`{
		"balance": 10.53,
		"transactions": [
			{
				"timestamp": "2020-10-23T14:05:01.199Z",
				"toAddress": "01234abcde",
				"amount": 20.97
			},
			{
				"timestamp": "2020-10-24T09:29:51.320Z",
				"fromAddress": "01234abcde",
				"toAddress": "98765zyxwt",
				"amount": 10.46
			}
		]
	}`)
	client := newClientMock(http.StatusOK, mockResponseBody, nil)
	jl := &JobcoinLib{client}

	_, err := jl.GetAddressInfo("01234abcde")
	if err == nil {
		t.Errorf("Expected error to be returned but it was not.")
	}

	assert.Contains(t, err.Error(), "json: cannot unmarshal")
}

// Begin SendJobcoin tests
func TestSendJobcoin_SendsJobcoinFromOneAddressToAnother(t *testing.T) {
	mockResponseBody := []byte(`
		{
			"status": "OK"
		}
	`)
	client := newClientMock(http.StatusOK, mockResponseBody, nil)
	jl := &JobcoinLib{client}

	err := jl.SendJobcoin("1234abcd", "9876zyxw", "11.23")
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}
}

func TestSendJobcoin_ReturnsErrorIfClientRequestFails(t *testing.T) {
	expectedErr := errors.New("request failed")
	client := newClientMock(0, nil, expectedErr)
	jl := &JobcoinLib{client}

	err := jl.SendJobcoin("1234abcd", "9876zyxw", "11.23")
	if err == nil {
		t.Errorf("Expected error to be returned but it was not.")
	}

	assert.Equal(t, expectedErr, err)
}

func TestSendJobcoin_ReturnsErrorIfTransactionCreationFails(t *testing.T) {
	mockResponseBody := []byte(`{
		"error": "Insufficient Funds"
	}`)
	client := newClientMock(http.StatusUnprocessableEntity, mockResponseBody, nil)
	jl := &JobcoinLib{client}

	err := jl.SendJobcoin("1234abcd", "9876zyxw", "11.23")
	if err == nil {
		t.Errorf("Expected error to be returned but it was not.")
	}

	expectedErr := "Failed to create transaction due to: Insufficient Funds"
	assert.Equal(t, expectedErr, err.Error())
}
