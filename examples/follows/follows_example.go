package main

import (
	"encoding/json"

	farcaster "github.com/ertan/go-farcaster/pkg"
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
	fc := farcaster.NewFarcasterClient(apiUrl, mnemonic, providerWs)
	println("Farcaster client created")

	users, _, err := fc.Follows.GetFollowingByFname("ertan", 10, "")
	if err != nil {
		panic(err)
	}
	prettyPrint(&users)
	users, _, err = fc.Follows.GetFollowersByFname("ertan", 10, "")
	if err != nil {
		panic(err)
	}
	prettyPrint(&users)
}
