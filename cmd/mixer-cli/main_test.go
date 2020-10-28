package main

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ckaminer/jobcoin/clientlib"
	"github.com/stretchr/testify/assert"
)

// Begin createMixerUser tests
func TestCreateMixerUser_ReturnsCreatedUser(t *testing.T) {
	mockResponseBody := []byte(`
		{
			"depositAddress": "deposit-one",
			"returnAddresses": [
				"return-one",
				"return-two",
				"return-three"
			]
		}
	`)
	client := clientlib.NewClientMock(http.StatusCreated, mockResponseBody, nil)

	returnAddresses := []string{"return-one", "return-two", "return-three"}
	createdUser, err := createMixerUser(client, returnAddresses)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.Equal(t, "deposit-one", createdUser.DepositAddress)
	assert.Equal(t, returnAddresses, createdUser.ReturnAddresses)
}

func TestCreateMixerUser_ReturnsErrorIfRequestFails(t *testing.T) {
	client := clientlib.NewClientMock(0, nil, errors.New("Request failed"))

	_, err := createMixerUser(client, nil)
	if err == nil {
		t.Errorf("Expected an error but did not receive one")
	}

	assert.Equal(t, "Request failed", err.Error())
}

func TestCreateMixerUser_ReturnsErrorIfFailsToDecodeError(t *testing.T) {
	mockResponseBody := []byte(`
		{
			"error": 100
		}
	`)
	client := clientlib.NewClientMock(http.StatusConflict, mockResponseBody, nil)

	returnAddresses := []string{"return-one", "return-two", "return-three"}
	_, err := createMixerUser(client, returnAddresses)
	if err == nil {
		t.Errorf("Expected an error but did not receive one")
	}

	assert.Contains(t, err.Error(), "cannot unmarshal")
}

func TestCreateMixerUser_ReturnsErrorIfFailsToDecodeUserResponse(t *testing.T) {
	mockResponseBody := []byte(`
		{
			"depositAddress": 100,
			"returnAddresses": [
				"return-one",
				"return-two",
				"return-three"
			]
		}
	`)
	client := clientlib.NewClientMock(http.StatusCreated, mockResponseBody, nil)

	returnAddresses := []string{"return-one", "return-two", "return-three"}
	_, err := createMixerUser(client, returnAddresses)
	if err == nil {
		t.Errorf("Expected an error but did not receive one")
	}

	assert.Contains(t, err.Error(), "cannot unmarshal")
}
