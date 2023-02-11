package account

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type AccountService struct {
	privateKey  *ecdsa.PrivateKey
	apiUrl      string
	accessToken string
	expiresAt   int64
	clock       func() time.Time
}

func (a *AccountService) now() time.Time {
	if a.clock == nil {
		return time.Now() // default implementation which fall back to standard library
	}

	return a.clock()
}

func NewAccountService(apiUrl, mnemonic string) *AccountService {
	if mnemonic == "" {
		// Currently the documentation requires access token for every endpoint but practically
		// some endpoints don't require it. So we can use a nil private key to skip auth.
		log.Print("No mnemonic provided, you might not be able to access private endpoints")
		return &AccountService{
			apiUrl: apiUrl,
		}
	}

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}
	account, err := wallet.Derive(hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0"), true)
	if err != nil {
		log.Fatal(err)
	}
	hexPrivateKey, err := wallet.PrivateKeyHex(account)
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA(hexPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	return &AccountService{
		privateKey: privateKey,
		apiUrl:     apiUrl,
	}
}

func (a *AccountService) GetAccessToken(expirationInSecs int) (string, error) {
	if a.privateKey == nil {
		return "", errors.New("private key is nil")
	}

	timestamp := a.now().UnixMilli()
	expiration := timestamp + int64(expirationInSecs*1000)
	if timestamp < a.expiresAt {
		return a.accessToken, nil
	}

	// Payload format
	type Params struct {
		ExpiresAt int64 `json:"expiresAt"`
		Timestamp int64 `json:"timestamp"`
	}
	type Payload struct {
		Method string `json:"method"`
		Params Params `json:"params"`
	}
	type Error struct {
		Message string `json:"message"`
	}
	type AccessToken struct {
		Secret string `json:"secret"`
	}
	type AccessTokenResponse struct {
		Result struct {
			Token AccessToken `json:"token"`
		} `json:"result"`
		Errors []Error `json:"errors"`
	}

	payload := Payload{
		Method: "generateToken",
		Params: Params{
			ExpiresAt: expiration,
			Timestamp: timestamp,
		},
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// EIP-191 spec: https://eips.ethereum.org/EIPS/eip-191
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(payloadJson), payloadJson)
	signHash := crypto.Keccak256Hash([]byte(msg)).Bytes()

	sig, err := crypto.Sign(signHash, a.privateKey)
	if err != nil {
		return "", err
	}

	base64Sig := base64.StdEncoding.EncodeToString(sig)
	bearer := fmt.Sprintf("eip191:%s", base64Sig)

	url := fmt.Sprintf("%s/v2/auth", a.apiUrl)
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payloadJson))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var response AccessTokenResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return "", err
	}
	if len(response.Errors) > 0 {
		return "", errors.New(response.Errors[0].Message)
	}
	a.accessToken = response.Result.Token.Secret
	a.expiresAt = expiration
	return a.accessToken, nil
}

func (a *AccountService) SendRequest(method, path string, params map[string]interface{}, body []byte) ([]byte, error) {
	url := a.apiUrl + path
	if len(params) > 0 {
		// TODO(ertan): Is this the best way to stringify params?
		url += "?"
		for key, value := range params {
			if _, ok := value.(string); ok {
				url += fmt.Sprintf("%s=%s&", key, value)
			} else if _, ok := value.(int); ok {
				url += fmt.Sprintf("%s=%d&", key, value)
			} else if _, ok := value.(uint64); ok {
				url += fmt.Sprintf("%s=%d&", key, value)
			} else {
				return nil, errors.New("params must be string or int")
			}
		}
		url = url[:len(url)-1]
	}
	// Create a new request using http
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if a.privateKey != nil {
		// Add auth header
		token, err := a.GetAccessToken(3600)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
	}
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}
