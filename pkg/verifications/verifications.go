package verifications

import (
	"encoding/json"
	"errors"

	"github.com/ertan/farcaster-go/pkg/account"
	"github.com/ertan/farcaster-go/pkg/registry"
	"github.com/ertan/farcaster-go/pkg/users"
)

type VerificationsService struct {
	account  *account.AccountService
	registry *registry.RegistryService
}

type Verification struct {
	Fid       uint64 `json:"fid"`
	Address   string `json:"address"`
	Timestamp uint64 `json:"timestamp"`
}

func NewVerificationsService(account *account.AccountService, registry *registry.RegistryService) *VerificationsService {
	return &VerificationsService{
		account:  account,
		registry: registry,
	}
}

func (v *VerificationsService) GetVerificationsByFid(fid int) ([]Verification, error) {
	type VerificationsResponse struct {
		Result struct {
			Verifications []Verification `json:"verifications"`
		} `json:"result"`
	}
	params := map[string]interface{}{
		"fid": fid,
	}

	responseBytes, err := v.account.SendRequest("GET", "/v2/verifications", params, nil)
	if err != nil {
		return nil, err
	}
	var response VerificationsResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		return response.Result.Verifications, nil
	}
	return nil, errors.New("Error getting verifications")
}

func (v *VerificationsService) GetUserByVerification(address string) (*users.User, error) {
	type UserResponse struct {
		Result struct {
			User users.User `json:"user"`
		} `json:"result"`
	}
	params := map[string]interface{}{
		"address": address,
	}

	responseBytes, err := v.account.SendRequest("GET", "/v2/user-by-verification", params, nil)
	if err != nil {
		return nil, err
	}
	var response UserResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		return &response.Result.User, nil
	}
	return nil, errors.New("Error getting user")
}
