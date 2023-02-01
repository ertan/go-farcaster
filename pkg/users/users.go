package users

import (
	"encoding/json"
	"errors"

	"github.com/ertan/farcaster-go/pkg/account"
	"github.com/ertan/farcaster-go/pkg/registry"
)

type UserService struct {
	account  *account.AccountService
	registry *registry.RegistryService
}

type User struct {
	Fid              int            `json:"fid"`
	Username         string         `json:"username"`
	DisplayName      string         `json:"displayName"`
	Pfp              Pfp            `json:"pfp"`
	Profile          Profile        `json:"profile"`
	FollowerCount    int            `json:"followerCount"`
	FollowingCount   int            `json:"followingCount"`
	ReferrerUsername string         `json:"referrerUsername"`
	ViewerContext    *ViewerContext `json:"viewerContext"`
}

type Profile struct {
	Bio Bio `json:"bio"`
}

type Bio struct {
	Text     string   `json:"text"`
	Mentions []string `json:"mentions"`
}

type Pfp struct {
	Url      string `json:"url"`
	Verified bool   `json:"verified"`
}

type ViewerContext struct {
	Following         bool `json:"following"`
	FollowedBy        bool `json:"followedBy"`
	CanSendDirectCast bool `json:"canSendDirectCast"`
}

func NewUserService(account *account.AccountService, registry *registry.RegistryService) *UserService {
	return &UserService{
		account:  account,
		registry: registry,
	}
}

type UserResponse struct {
	Result *User `json:"result"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (s *UserService) getUser(path string, params map[string]interface{}) (*User, error) {
	responseBytes, err := s.account.SendRequest("GET", path, params, nil)
	if err != nil {
		return nil, err
	}
	var response UserResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, errors.New(response.Errors[0].Message)
		}
		return response.Result, nil
	}
	return nil, err
}

func (u *UserService) GetUserByFid(fid uint64) (*User, error) {
	return u.getUser("/v2/user", map[string]interface{}{"fid": fid})
}

func (u *UserService) GetUserByUsername(username string) (*User, error) {
	return u.getUser("/v2/user-by-username", map[string]interface{}{"username": username})
}

func (u *UserService) GetUserByAddress(address string) (*User, error) {
	if u.registry == nil {
		return nil, errors.New("Registry service is not initialized. Use GetUserByFid or GetUserByUsername.")
	}
	fid, err := u.registry.GetFidByAddress(address)
	if err != nil {
		return nil, err
	}
	return u.GetUserByFid(fid)
}

func (u *UserService) getCustodyAddress(params map[string]interface{}) (string, error) {
	responseBytes, err := u.account.SendRequest("GET", "/v2/custody-address", params, nil)
	if err != nil {
		return "", err
	}
	var Response struct {
		Result struct {
			CustodyAddress string `json:"custodyAddress"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(responseBytes, &Response); err == nil {
		if len(Response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return "", errors.New(Response.Errors[0].Message)
		}
		return Response.Result.CustodyAddress, nil
	}
	return "", err
}

func (u *UserService) GetCustodyAddressByFid(fid uint64) (string, error) {
	return u.getCustodyAddress(map[string]interface{}{"fid": fid})
}

func (u *UserService) GetCustodyAddressByUsername(username string) (string, error) {
	return u.getCustodyAddress(map[string]interface{}{"fname": username})
}

func (u *UserService) Me() (*User, error) {
	return u.getUser("/v2/me", nil)
}

func (u *UserService) GetRecentUsers(limit int, cursor string) ([]User, string, error) {
	params := map[string]interface{}{}
	if limit > 0 {
		params["limit"] = limit
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	responseBytes, err := u.account.SendRequest("GET", "/v2/recent-users", params, nil)
	if err != nil {
		return nil, "", err
	}
	var response struct {
		Result struct {
			Users []User `json:"users"`
		} `json:"result"`
		// The documentation does not mention a cursor but the example response has it.
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Users, response.Next.Cursor, nil
	}
	return nil, "", err
}
