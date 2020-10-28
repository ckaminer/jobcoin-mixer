# Golang Jobcoin Mixer
This repository is based off of the Golang Jobcoin boilerplate which can be found [here](https://github.com/gemini/jobcoin-boilerplate/tree/master/golang).

This application works to mix your Jobcoin to increase anonymity on the Jobcoin network.

## How it works

1. Using the CLI or API, provide a list of unused addresses that you own.
2. You will receive a deposit address in return. Whenever you are ready, send your Jobcoin to this address.
3. Once the mixer detects this deposit, it will "mix" it by sending it to a house account. Upon doing so, a fee will also be collected by the mixer.
4. Once your Jobcoin reaches the house account, it will be returned back to your addresses provided in step 1, over time,  in small, random amounts.

## About
This project contains an API that is used to mix Jobcoin. The api accepts new Jobcoin users via the endpoint listed below. If you prefer, there is also a small CLI that also requires return addresses as an input but will handle the API interaction for you.

## Usage
### Clean + Deps + Build
```
make all
```

### Clean
```
make clean
```

### Deps
```
make deps
```

### Test
```
make test
```

### API
The application is configured to run on port **8080**. If you would like to run it on a different port you may do so by updating the `MixerPort` variable in `config.go`.

To build the app:
```
make build-api
```

To run the app:
```
./bin/mixer-api
```

#### Endpoints
- Create User

  `POST api/users`

  Sample Request Body:
  ```
  {
    "returnAddresses": [
      "how",
      "now",
      "brown",
      "cow"
    ]
  }
  ```

  Expected Response:
  ```
  {
    "depositAddress": "23fa4cfe-194a-11eb-a23d-f45c8995c541",
    "returnAddresses": [
      "how",
      "now",
      "brown",
      "cow"
    ]
  }
  ```
#### Manual Testing Conigurations
If you would like to test the application by hand, there are a couple of configurations you may wish to temporarily change.

- To alter the timing interval during polling (for new users and for house users) you can update the time being passed into the tickers created in `./cmd/mixer-api/main.go#main`.

- The house address is currently reset every time the app is started. This is by design due to the ephemeral nature of the application design. In future, long-term iterations, this would be hidden and consistent. However, if you would like to keep it consistent you may comment out the following line in `./cmd/mixer-api/main.go#main`:
  ```
  mixerlib.HouseAddress = houseAddress.String()
  ```
  Doing this will cause the HouseAddress to use its default value which can be found/changed in `./mixerlib/lib.go`.

- The `MixerBankFund` is the address that collected service fees will be sent to. If you would like to change this it can also be found in `./mixerlib/lib.go`.

- The service fee being collected is currently `1%` per deposit, or a multiplier of `0.01`. To change this value, you can update `ServiceFeePctg` in `./mixerlib/lib.go`.

- In an effort to remain conspicuous, the mixer will only return up to 5 Jobcoin at a time back to your user-provided return addresses. If you would like to change this amount you may do so by changing the following value in `./mixerlib/lib.go`:
  ```
  const DistributionIncrement = 5.0
  ```

### CLI
Note: If you prefer to use the CLI to create your Jobcoin Mixer deposit address, please be sure the API is already running.

Once the API is running, in a new tab, you may execute the following commands:

Build the CLI
```
make build-cli
```

Run the CLI
```
./bin/mixer-cli --addresses=how,now,brown,cow
```

Expected output
```
You may now send Jobcoins to address 0bd1c388-1944-11eb-848f-f45c8995c541.

They will be mixed into [how now brown cow] and sent to your destination addresses.
```

### Tracking Your Funds
Upon creation of your Mixer User you should receive a deposit address from either the API response or the CLI output which can both be found above. Once you have your deposit address you are free to start mixing! Using the [Jobcoin UI](https://jobcoin.gemini.com/casino-unit) you may begin by sending Jobcoin from any address to your deposit address. Once that is complete, depending on how much Jobcoin you sent, you need to do nothing but wait for your Jobcoin to be returned back to you. If your deposit is less than the 5.0 Jobcoin distribution increment this process should take no more than 15 seconds. Once enough time has passed, there should be a few transactions that you can check (either via the Jobcoin UI linked above or the [transactions endpoint](http://jobcoin.gemini.com/casino-unit/api/transactions
)) to verify that the Mixer is working properly. You should be able to see:

1. The transaction you created when you sent Jobcoin *from* the address of your choosing *to* your Mixer deposit address.
2. There should be a transaction *from* your deposit address *to* the house address. For reference, the house address is printed out to the server logs immediately upon startup of the api.
3. There should be a series of transactions *from* the house address *to* the return addresses you provided to the Mixer when you created your user. The *sum* of these transactions should equal your initial deposit. 

## Development Diary
#### Challenging Design Decisions
- When I first started sketching out my plan, I was going to track transactions made by the mixer on each individual user in memory i.e. `user.DepositTxs`, `user.HouseTxs`, `user.ReturnTxs`
  - These would be used to determine the flow and balance of each user money from origin -> deposit -> house -> return
  - I shortly thereafter decided this would be better suited for a longer-term solution where potentially a database could be involved
  - I also realized storing entire transactions wasn't necessary and that the same thing could be accomplished  instead by storing and maintaining strictly balances
    - i.e. - Send money from deposit address to house address, increase the `user.HouseBalance`. Send money from house to return address, decrease the `user.HouseBalance`
    - This created additional complexity with higher risk of error
      - It was difficult to maintain balances while there were two asynchronous processes happening that could require changes to the same user
      - Before moving on to my final solution I attempted to switch over to passing around pointers to users instead of references. This alleviated some of the initial painpoints but sent me down a path where there was a power struggle in application design. There was a half-FP half-OOP design that I wanted to avoid in order to maintain consistency and readability.
  - Ultimately I decided to leverage the Jobcoin API to calculate balances of user money in the house
    - The main tradeoff here is the sacrifice of speed for accuracy. Leaning on the API as the only source of record ensures that nothing was mixed up within the application along the way. The drawback is that pulling House transactions will eventually return a massive payload that takes a lot of time and resources to filter through. For now, with the small scale of the mixer, I think the accuracy is more than worth it - the last thing I want to do is return incorrect amounts of money to users.

- Re-iterating a point made in the Manual Testing Configurations section above, due to the ephemeral nature of the app, I decided to make the house address something that is reset each time the app runs. Because there is no database and Mixer Users are stored in memory, there can be issues when the app is run, used, stopped, re-started, and then used again with the same return addresses as the first run. User house balances are calculated based off of transactions returned from the API so when a "different" user uses the same return addresses the return amounts can be mis-calculated. User return addresses are validated for uniqueness upon user creation but this validation can only track users created within one particular run of the app. However, the Jobcoin transactions are of course eternal so the house address needs to be reset each app run to avoid mis-calculations.

- Ideally I would love to offload the sorting of transactions to someplace other than the Go app - either through filtering on the API or a database that stores only transactions created by the Mixer. This would resolve the issue above regarding the house address and would similarly make the process of determing the location of each user's money much faster.

- Collecting a fee upon the transfer of funds from the deposit address to the house address instead of upon the return of funds back to the user was a decision that was made for a couple reasons. For one, it decreases the amount of transactions being created by the app. Due to the reasons described above (i.e. filtering through transactions in the app), I want to create as few transactions as possible. It also should assist with spending less money on on-chain transaction fees (these don't exist in Jobcoin but would otherwise). The drawback is that it could put the maintainers of the Jobcoin Mixer (myself) in an uncomfortable scenario if the Mixer were to collect a fee then have some sort of bug that prevents users and/or house from receiving their money. A total PR nightmare.