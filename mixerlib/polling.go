package mixerlib

import (
	"log"
	"time"
)

// PollForNewDeposits is an endlessly looping function handling the input of new users.
// When a user comes in through the provided user channel they are added to MixerUsers.
// On a steady time interval each MixerUser is passed to transferDepositToHouse to
// potentially move funds if necessary.
func (ml *MixerLib) PollForNewDeposits(ticker *time.Ticker, userChan, houseChan chan MixerUser) {
	for {
		select {
		case <-ticker.C:
			for _, user := range MixerUsers {
				sentToHouse, _ := ml.transferDepositToHouse(user)
				if sentToHouse {
					houseChan <- user
				}
			}
		case newUser := <-userChan:
			log.Printf("Adding user %s to MixerUsers", newUser.DepositAddress)
			MixerUsers = append(MixerUsers, newUser)
		}
	}
}

// PollForUserReturns is an endlessly looping function handling the redistribution of money.
// When a user comes in through the provided house channel they are added to the
// HouseQueue. On a steady time interval each user in the queue will have some of their
// funds returned back to them.
func (ml *MixerLib) PollForUserReturns(ticker *time.Ticker, houseChan chan MixerUser) {
	for {
		select {
		case <-ticker.C:
			usersStillInHouse := []MixerUser{}
			for _, user := range HouseQueue {
				emptyBalance, _ := ml.returnFundsToUser(user)
				if !emptyBalance {
					usersStillInHouse = append(usersStillInHouse, user)
				}
			}
			HouseQueue = usersStillInHouse
		case houseUser := <-houseChan:
			log.Printf("Adding user %s to HouseQueue", houseUser.DepositAddress)
			HouseQueue = addOrReplaceUserInCollection(HouseQueue, houseUser)
		}
	}
}
