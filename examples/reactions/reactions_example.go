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
	farcaster := farcaster.NewFarcasterClient(apiUrl, mnemonic, providerWs)
	println("Farcaster client created")

	reactions, _, err := farcaster.Reactions.GetReactionsByCastHash("0x8b20fdcbf77255770400da9a50a9c31ba151337fee7ab19a6d337b6da7744769", 0, "")
	if err != nil {
		panic(err)
	}
	prettyPrint(&reactions)
	users, _, err := farcaster.Reactions.GetRecastersByCastHash("0x321712dc8eccc5d2be38e38c1ef0c8916c49949a80ffe20ec5752bb23ea4d86f", 0, "")
	if err != nil {
		panic(err)
	}
	prettyPrint(&users)
	reactions, _, err = farcaster.Reactions.GetUserReactions(40)
	if err != nil {
		panic(err)
	}
	prettyPrint(&reactions)

	// Uncomment for mutating examples
	// reaction, err := farcaster.Reactions.ReactToCast("0x24b6f70d58ca9bd48f2372d329f3ba0ed6d569550698928b7bf00897e5a6d19a")
	// if err != nil {
	// 	panic(err)
	// }
	// prettyPrint(&reaction)
	// err = farcaster.Reactions.UnreactToCast("0x24b6f70d58ca9bd48f2372d329f3ba0ed6d569550698928b7bf00897e5a6d19a")
	// if err != nil {
	// 	panic(err)
	// }
	// castHash, err := farcaster.Reactions.RecastCast("0x24b6f70d58ca9bd48f2372d329f3ba0ed6d569550698928b7bf00897e5a6d19a")
	// if err != nil {
	// 	panic(err)
	// }
	// prettyPrint(&castHash)
	// err = farcaster.Reactions.UnrecastCast("0x24b6f70d58ca9bd48f2372d329f3ba0ed6d569550698928b7bf00897e5a6d19a")
	// if err != nil {
	// 	panic(err)
	// }
}
