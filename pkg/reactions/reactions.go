package reactions

import (
	"encoding/json"
	"errors"

	"github.com/ertan/go-farcaster/pkg/account"
	"github.com/ertan/go-farcaster/pkg/users"
)

type ReactionService struct {
	account *account.AccountService
}

type Reaction struct {
	Type      string      `json:"type"`
	Hash      string      `json:"hash"`
	Reactor   *users.User `json:"reactor"`
	Timestamp uint64      `json:"timestamp"`
	CastHash  string      `json:"castHash"`
}

func NewReactionService(account *account.AccountService) *ReactionService {
	return &ReactionService{
		account: account,
	}
}

func (r *ReactionService) GetReactionsByCastHash(hash string, limit int, cursor string) ([]Reaction, string, error) {
	type ReactionsResponse struct {
		Result struct {
			// This is called `likes` not `reactions` in the API
			Reactions []Reaction `json:"likes"`
		} `json:"result"`
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	params := map[string]interface{}{
		"castHash": hash,
	}
	if limit > 0 {
		params["limit"] = limit
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	responseBytes, err := r.account.SendRequest("GET", "/v2/cast-likes", params, nil)
	if err != nil {
		return nil, "", err
	}
	println(string(responseBytes))
	var response ReactionsResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Reactions, response.Next.Cursor, nil
	}
	return nil, "", errors.New("Error fetching reactions")
}

func (r *ReactionService) ReactToCast(hash string) (*Reaction, error) {
	type ReactionRequest struct {
		CastHash string `json:"castHash"`
	}
	type ReactionResponse struct {
		Result struct {
			Reaction *Reaction `json:"like"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	request := ReactionRequest{
		CastHash: hash,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	responseBytes, err := r.account.SendRequest("PUT", "/v2/cast-likes", nil, requestBytes)
	if err != nil {
		return nil, err
	}
	var response ReactionResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, errors.New(response.Errors[0].Message)
		}
		return response.Result.Reaction, nil
	}
	return nil, errors.New("Error reacting to cast")
}

func (r *ReactionService) UnreactToCast(hash string) error {
	type ReactionRequest struct {
		CastHash string `json:"castHash"`
	}
	type ReactionResponse struct {
		Result struct {
			Success bool `json:"success"`
		} `json:"result"`
	}
	request := ReactionRequest{
		CastHash: hash,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	responseBytes, err := r.account.SendRequest("DELETE", "/v2/cast-likes", nil, requestBytes)
	if err != nil {
		return err
	}
	var response ReactionResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if response.Result.Success {
			return nil
		}
	}
	return errors.New("Error unreacting to cast")
}

func (r *ReactionService) GetRecastersByCastHash(hash string, limit int, cursor string) ([]users.User, string, error) {
	type RecastersResponse struct {
		Result struct {
			Recasters []users.User `json:"users"`
		} `json:"result"`
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	params := map[string]interface{}{
		"castHash": hash,
	}
	if limit > 0 {
		params["limit"] = limit
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	responseBytes, err := r.account.SendRequest("GET", "/v2/cast-recasters", params, nil)
	if err != nil {
		return nil, "", err
	}
	var response RecastersResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Recasters, response.Next.Cursor, nil
	}
	return nil, "", errors.New("Error fetching recasters")
}

func (r *ReactionService) RecastCast(hash string) (string, error) {
	type ReactionRequest struct {
		CastHash string `json:"castHash"`
	}
	type ReactionResponse struct {
		Result struct {
			CastHash string `json:"castHash"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	request := ReactionRequest{
		CastHash: hash,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	responseBytes, err := r.account.SendRequest("PUT", "/v2/recasts", nil, requestBytes)
	if err != nil {
		return "", err
	}
	var response ReactionResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return "", errors.New(response.Errors[0].Message)
		}
		return response.Result.CastHash, nil
	}
	return "", errors.New("Error recasting cast")
}

func (r *ReactionService) UnrecastCast(hash string) error {
	type ReactionRequest struct {
		CastHash string `json:"castHash"`
	}
	type ReactionResponse struct {
		Result struct {
			Success bool `json:"success"`
		} `json:"result"`
	}
	request := ReactionRequest{
		CastHash: hash,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	responseBytes, err := r.account.SendRequest("DELETE", "/v2/recasts", nil, requestBytes)
	if err != nil {
		return err
	}
	var response ReactionResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if response.Result.Success {
			return nil
		}
	}
	return errors.New("Error unrecasting cast")
}

func (r *ReactionService) GetUserReactions(fid uint64) ([]Reaction, string, error) {
	type UserReactionsResponse struct {
		Result struct {
			Reactions []Reaction `json:"likes"`
		} `json:"result"`
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	params := map[string]interface{}{
		"fid": fid,
	}
	responseBytes, err := r.account.SendRequest("GET", "/v2/user-cast-likes", params, nil)
	if err != nil {
		return nil, "", err
	}
	var response UserReactionsResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Reactions, response.Next.Cursor, nil
	}
	return nil, "", errors.New("Error fetching user reactions")
}
