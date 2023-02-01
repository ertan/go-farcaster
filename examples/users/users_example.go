package main

import (
	"encoding/json"
	"log"

	farcaster "github.com/ertan/farcaster-go/pkg"
	"github.com/spf13/viper"
)

func prettyPrint(st interface{}) {
	stJson, err := json.Marshal(st)
	if err != nil {
		panic(err)
	}
	println(string(stJson))
}

func main() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	apiUrl := viper.Get("FARCASTER_API_URL").(string)
	mnemonic := viper.Get("FARCASTER_MNEMONIC").(string)
	providerWs := viper.Get("ETHEREUM_PROVIDER_WS").(string)
	fc := farcaster.NewFarcasterClient(apiUrl, mnemonic, "")
	println("Farcaster client created")

	user, err := fc.Users.GetUserByFid(40)
	if err != nil {
		panic(err)
	}
	prettyPrint(&user)
	user, err = fc.Users.GetUserByUsername("ertan")
	if err != nil {
		panic(err)
	}
	prettyPrint(&user)

	custodyAddress, err := fc.Users.GetCustodyAddressByUsername("ertan")
	if err != nil {
		panic(err)
	}
	println(custodyAddress)

	user, err = fc.Users.GetUserByAddress(custodyAddress)
	if err != nil {
		log.Println(err)
	}

	fc = farcaster.NewFarcasterClient(apiUrl, mnemonic, providerWs)
	println("New Farcaster client with registry is created")
	user, err = fc.Users.GetUserByAddress(custodyAddress)
	if err != nil {
		panic(err)
	}
	prettyPrint(&user)

	user, err = fc.Users.Me()
	if err != nil {
		panic(err)
	}
	prettyPrint(&user)

	users, _, err := fc.Users.GetRecentUsers(10, "")
	if err != nil {
		panic(err)
	}
	prettyPrint(&users)

	// Verification examples
	// verifications, err := fc.Verifications.GetVerificationsByFid(40)
	// if err != nil {
	// 	panic(err)
	// }
	// prettyPrint(&verifications)

	// user, err = fc.Verifications.GetUserByVerification("")
	// if err != nil {
	// 	panic(err)
	// }
	// prettyPrint(&user)
}
