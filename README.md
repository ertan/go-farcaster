# Farcaster Go Client
This is the Golang implementation for Farcaster client based on the [official documentation](https://farcasterxyz.notion.site/Merkle-v2-API-Documentation-c19a9494383a4ce0bd28db6d44d99ea8#a30b0e73495240e1a091a9cd5b736fb8).

Inspired by [Rust](https://github.com/TheLDB/farcaster-rs) and [Python](https://github.com/a16z/farcaster-py) client implementations. Coauthored by GitHub Copilot and ChatGPT. 🙏

## Prerequisites
In order to test the examples you need to set the following environment variables in .env file.
```
FARCASTER_API_URL    = "https://api.farcaster.xyz"
FARCASTER_MNEMONIC   = "your mnemonic"
ETHEREUM_PROVIDER_WS = "your Goerli endpoint"
```
Registry is built based on the event logs to get fid <> fname <> address mappings. If `ETHEREUM_PROVIDER_WS` variable isn't set, you can still use the API. Mnemonic is required for authorization to access most of the API endpoints. However, it's not required by the client as some endpoints are open to public.

## Examples
Some examples to test the client are:
```
go run examples/casts/casts_example.go
```
```
go run examples/reactions/reactions_example.go
```
```
go run examples/users/users_example.go
```
You can find other example runs under `examples/` directory.

## Future Work
- Tests! There are currently no unit tests for the client, just examples. 😅
- Missing comments on exported functions and structs. 
