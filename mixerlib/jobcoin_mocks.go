package mixerlib

import (
	"github.com/ckaminer/jobcoin/clientlib"
)

type mockJobcoinClient struct {
	AddressInfo     clientlib.JobcoinAddressInfo
	GetAddressError error
	SendError       error
}

func (mc *mockJobcoinClient) GetAddressInfo(address string) (clientlib.JobcoinAddressInfo, error) {
	return mc.AddressInfo, mc.GetAddressError
}

func (mc *mockJobcoinClient) SendJobcoin(fromAddress, toAddress, amount string) error {
	return mc.SendError
}

// newJobcoinMock returns a mock implementation of the JobcoinClient interface.
// addressInfo will be the value returned in GetAddressInfo.
// addressErr will be the error returned in GetAddressInfo.
// sendErr will be the error returned in SendJobcoin.
func newJobcoinMock(addressInfo clientlib.JobcoinAddressInfo, addressErr, sendErr error) clientlib.JobcoinClient {
	return &mockJobcoinClient{
		AddressInfo:     addressInfo,
		GetAddressError: addressErr,
		SendError:       sendErr,
	}
}
