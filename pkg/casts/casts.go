package casts

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ertan/farcaster-go/pkg/account"
	"github.com/ertan/farcaster-go/pkg/registry"
	"github.com/ertan/farcaster-go/pkg/users"
)

type CastService struct {
	account  *account.AccountService
	registry *registry.RegistryService
}

type Cast struct {
	Hash          string         `json:"hash"`
	ThreadHash    string         `json:"threadHash"`
	ParentHash    string         `json:"parentHash"`
	ParentAuthor  *users.User    `json:"parentAuthor"`
	Author        *users.User    `json:"author"`
	Text          string         `json:"text"`
	Timestamp     uint64         `json:"timestamp"`
	Replies       *Replies       `json:"replies"`
	Reactions     *Reactions     `json:"reactions"`
	Recasts       *Recasts       `json:"recasts"`
	Watches       *Watches       `json:"watches"`
	Recast        bool           `json:"recast"`
	ViewerContext *ViewerContext `json:"viewerContext"`
}

type Replies struct {
	Count int `json:"count"`
}

type Reactions struct {
	Count int `json:"count"`
}

type Recasts struct {
	Count   int         `json:"count"`
	Recasts []*Recaster `json:"recasters"`
}

type Recaster struct {
	Fid         uint64 `json:"fid"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	RecastHash  string `json:"recastHash"`
}

type Watches struct {
	Count int `json:"count"`
}

type ViewerContext struct {
	Reacted bool `json:"reacted"`
	Recast  bool `json:"recast"`
	Watched bool `json:"watched"`
}

type CastsResponse struct {
	Result struct {
		Casts []Cast `json:"casts"`
	} `json:"result"`
	Next struct {
		Cursor string `json:"cursor"`
	} `json:"next"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func NewCastService(account *account.AccountService, registry *registry.RegistryService) *CastService {
	return &CastService{
		account:  account,
		registry: registry,
	}
}

func (c *CastService) GetCastByHash(hash string) (*Cast, error) {
	type CastResponse struct {
		Result struct {
			Cast *Cast `json:"cast"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	params := map[string]interface{}{
		"hash": hash,
	}
	responseBytes, err := c.account.SendRequest("GET", "/v2/cast", params, nil)
	if err != nil {
		return nil, err
	}
	var response CastResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, errors.New(response.Errors[0].Message)
		}
		return response.Result.Cast, nil
	}
	return nil, errors.New("Error fetching cast")
}

func (c *CastService) GetCastsByFid(fid uint64, limit int, cursor string) ([]Cast, string, error) {
	params := map[string]interface{}{
		"fid": fid,
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	responseBytes, err := c.account.SendRequest("GET", "/v2/casts", params, nil)
	if err != nil {
		return nil, "", err
	}
	var response CastsResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Casts, response.Next.Cursor, nil
	}
	return nil, "", errors.New("Error fetching casts")
}

func (c *CastService) GetCastsByFname(fname string, limit int, cursor string) ([]Cast, string, error) {
	if c.registry == nil {
		return nil, "", errors.New("Registry service is not initialized. Use GetCastsByFid.")
	}
	fid, err := c.registry.GetFidByFname(fname)
	if err != nil {
		return nil, "", err
	}
	return c.GetCastsByFid(fid, limit, cursor)
}

func (c *CastService) GetCastsByAddress(address string, limit int, cursor string) ([]Cast, string, error) {
	if c.registry == nil {
		return nil, "", errors.New("Registry service is not initialized. Use GetCastsByFid.")
	}
	fid, err := c.registry.GetFidByAddress(address)
	if err != nil {
		return nil, "", err
	}
	return c.GetCastsByFid(fid, limit, cursor)
}

func (c *CastService) GetCastsInThread(threadHash string) ([]Cast, error) {
	params := map[string]interface{}{
		"threadHash": threadHash,
	}
	responseBytes, err := c.account.SendRequest("GET", "/v2/all-casts-in-thread", params, nil)
	if err != nil {
		return nil, err
	}
	var response CastsResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, errors.New(response.Errors[0].Message)
		}
		return response.Result.Casts, nil
	}
	return nil, err
}

func (c *CastService) publishCast(request []byte) (*Cast, error) {
	type PublishCastResponse struct {
		Result struct {
			Cast *Cast `json:"cast"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	responseBytes, err := c.account.SendRequest("POST", "/v2/casts", nil, request)
	if err != nil {
		return nil, err
	}
	var response PublishCastResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}
	if len(response.Errors) > 0 {
		// TODO(ertan): Find a better solution to pass the errors here.
		return nil, errors.New(response.Errors[0].Message)
	}
	return response.Result.Cast, nil
}

func (c *CastService) PublishCast(text string) (*Cast, error) {
	type PublishCastRequest struct {
		Text string `json:"text"`
	}
	request := PublishCastRequest{
		Text: text,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	return c.publishCast(requestBytes)
}

func (c *CastService) PublishReplyCast(text string, fid uint64, hash string) (*Cast, error) {
	type Parent struct {
		Fid  uint64 `json:"fid"`
		Hash string `json:"hash"`
	}
	type PublishCastRequest struct {
		Text   string  `json:"text"`
		Parent *Parent `json:"parent"`
	}
	request := PublishCastRequest{
		Text: text,
		Parent: &Parent{
			Fid:  fid,
			Hash: hash,
		},
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	return c.publishCast(requestBytes)
}

func (c *CastService) DeleteCast(castHash string) error {
	type DeleteCastRequest struct {
		CastHash string `json:"castHash"`
	}
	type DeleteCastResponse struct {
		Result struct {
			Success bool `json:"success"`
		} `json:"result"`
	}
	request := DeleteCastRequest{
		CastHash: castHash,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	responseBytes, err := c.account.SendRequest("DELETE", "/v2/casts", nil, requestBytes)
	if err != nil {
		return err
	}
	var response DeleteCastResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return err
	}
	if !response.Result.Success {
		return errors.New("failed to delete cast")
	}
	return nil
}

func (c *CastService) GetRecentCasts(limit int) ([]Cast, string, error) {
	params := make(map[string]interface{})
	if limit > 0 {
		params["limit"] = limit
	}
	responseBytes, err := c.account.SendRequest("GET", "/v2/recent-casts", params, nil)
	if err != nil {
		return nil, "", err
	}
	var response CastsResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Casts, response.Next.Cursor, nil
	}
	return nil, "", err
}
