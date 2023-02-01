package follows

import (
	"encoding/json"
	"errors"

	"github.com/ertan/farcaster-go/pkg/account"
	"github.com/ertan/farcaster-go/pkg/registry"
	"github.com/ertan/farcaster-go/pkg/users"
)

type FollowService struct {
	account  *account.AccountService
	registry *registry.RegistryService
}

type Follow struct {
	FollowerFid  int64 `json:"follower_fid"`
	FollowingFid int64 `json:"following_fid"`
}

func NewFollowService(account *account.AccountService, registry *registry.RegistryService) *FollowService {
	return &FollowService{
		account:  account,
		registry: registry,
	}
}

func (f *FollowService) Follow(fid uint64) error {
	type FollowRequest struct {
		Fid uint64 `json:"targetFid"`
	}
	type FollowResponse struct {
		Result struct {
			Success bool `json:"success"`
		} `json:"result"`
	}
	request := FollowRequest{
		Fid: fid,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	responseBytes, err := f.account.SendRequest("PUT", "/v2/follows", nil, requestBytes)
	if err != nil {
		return err
	}
	var response FollowResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if response.Result.Success {
			return nil
		}
	}
	return errors.New("Error following user")
}

func (f *FollowService) Unfollow(fid uint64) error {
	type UnfollowRequest struct {
		Fid uint64 `json:"targetFid"`
	}
	type UnfollowResponse struct {
		Result struct {
			Success bool `json:"success"`
		} `json:"result"`
	}
	request := UnfollowRequest{
		Fid: fid,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	responseBytes, err := f.account.SendRequest("DELETE", "/v2/follows", nil, requestBytes)
	if err != nil {
		return err
	}
	var response UnfollowResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if response.Result.Success {
			return nil
		}
	}
	return errors.New("Error unfollowing user")
}

func (f *FollowService) getFollows(path string, fid uint64, limit int, cursor string) ([]users.User, string, error) {
	type FollowersResponse struct {
		Result struct {
			Users []users.User `json:"users"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
	}
	params := map[string]interface{}{
		"fid": fid,
	}
	if limit > 0 {
		params["limit"] = limit
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	responseBytes, err := f.account.SendRequest("GET", path, params, nil)
	if err != nil {
		return nil, "", err
	}
	var response FollowersResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Users, response.Next.Cursor, nil
	}
	return nil, "", errors.New("Error getting follows")
}

func (f *FollowService) GetFollowersByFid(fid uint64, limit int, cursor string) ([]users.User, string, error) {
	return f.getFollows("/v2/followers", fid, limit, cursor)
}

func (f *FollowService) GetFollowersByFname(fname string, limit int, cursor string) ([]users.User, string, error) {
	if f.registry == nil {
		return nil, "", errors.New("Registry service is not initialized. Use GetFollowersByFid.")
	}
	fid, err := f.registry.GetFidByFname(fname)
	if err != nil {
		return nil, "", err
	}
	return f.getFollows("/v2/followers", fid, limit, cursor)
}

func (f *FollowService) GetFollowingByFid(fid uint64, limit int, cursor string) ([]users.User, string, error) {
	return f.getFollows("/v2/following", fid, limit, cursor)
}

func (f *FollowService) GetFollowingByFname(fname string, limit int, cursor string) ([]users.User, string, error) {
	if f.registry == nil {
		return nil, "", errors.New("Registry service is not initialized. Use GetFollowingByFid.")
	}
	fid, err := f.registry.GetFidByFname(fname)
	if err != nil {
		return nil, "", err
	}
	return f.getFollows("/v2/following", fid, limit, cursor)
}
