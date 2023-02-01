package main

import (
	farcaster "github.com/ertan/farcaster-go/pkg"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	apiUrl := viper.Get("FARCASTER_API_URL").(string)
	mnemonic := viper.Get("FARCASTER_MNEMONIC").(string)
	providerWs := viper.Get("ETHEREUM_PROVIDER_WS").(string)
	farcaster := farcaster.NewFarcasterClient(apiUrl, mnemonic, providerWs)
	println("Farcaster client created")

	err := farcaster.Health.OK()
	if err != nil {
		panic(err)
	}
	println("Health check passed")
}
