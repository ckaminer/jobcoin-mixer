package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ckaminer/jobcoin/mixerlib"
	"github.com/stretchr/testify/assert"
)

func TestCreateNewUserHandler_ReturnsHandlerFuncForNewUsers(t *testing.T) {
	// Buffer the UserChannel so that it closes once single value comes through
	// For testing only to avoid hanging tests
	userChan := make(chan mixerlib.MixerUser, 1)
	newUserHandlerFunc := CreateNewUserHandler(userChan)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(newUserHandlerFunc)

	reqBody := []byte(`
		{
			"returnAddresses": [
				"return-one",
				"return-two",
				"return-three"
			]
		}
	`)

	r, _ := http.NewRequest("POST", "api/users", bytes.NewReader(reqBody))

	handler.ServeHTTP(recorder, r)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var resBody mixerlib.MixerUser
	err := json.NewDecoder(recorder.Body).Decode(&resBody)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.NotEqual(t, "", resBody.DepositAddress)
	assert.Equal(t, 36, len(resBody.DepositAddress))
	assert.Equal(t, "return-one", resBody.ReturnAddresses[0])
	assert.Equal(t, "return-two", resBody.ReturnAddresses[1])
	assert.Equal(t, "return-three", resBody.ReturnAddresses[2])
}

func TestCreateNewUserHandler_ReturnsConflictIfInvalidReturnAddress(t *testing.T) {
	newUserHandlerFunc := CreateNewUserHandler(nil)
	// Add users to MixerUser to create return address conflict
	mixerlib.MixerUsers = []mixerlib.MixerUser{
		{
			ReturnAddresses: []string{
				"return-two",
			},
		},
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(newUserHandlerFunc)

	reqBody := []byte(`
		{
			"depositAddress": "deposit-one",
			"returnAddresses": [
				"return-one",
				"return-two",
				"return-three"
			]
		}
	`)

	r, _ := http.NewRequest("POST", "api/users", bytes.NewReader(reqBody))

	handler.ServeHTTP(recorder, r)

	assert.Equal(t, http.StatusConflict, recorder.Code)

	var resBody ErrorPayload
	err := json.NewDecoder(recorder.Body).Decode(&resBody)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.Equal(t, "Return address return-two is already in use", resBody.Message)
}

func TestCreateNewUserHandler_ReturnsBadRequestIfInvalidReqBody(t *testing.T) {
	newUserHandlerFunc := CreateNewUserHandler(nil)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(newUserHandlerFunc)

	reqBody := []byte(`
		{
			"depositAddress": 1,
			"returnAddresses": [
				"return-one",
				"return-two",
				"return-three"
			]
		}
	`)

	r, _ := http.NewRequest("POST", "api/users", bytes.NewReader(reqBody))

	handler.ServeHTTP(recorder, r)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var resBody ErrorPayload
	err := json.NewDecoder(recorder.Body).Decode(&resBody)
	if err != nil {
		t.Errorf("Did not expect error. Got: %s", err.Error())
	}

	assert.Equal(t, "Invalid request body", resBody.Message)
}
