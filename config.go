package jobcoin

// Configuration defines a base minimum configuration for the jobcoin mixer
const (
	BaseURL             = "https://jobcoin.gemini.com/casino-unit/api"
	AddressesEndpoint   = BaseURL + "/addresses"
	TransactionEndpoint = BaseURL + "/transactions"
	MixerPort           = ":8080"
	MixerBaseURL        = "http://localhost" + MixerPort + "/api"
	MixerUserEndpoint   = MixerBaseURL + "/users"
)
