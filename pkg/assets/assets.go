package assets

import (
	"encoding/json"
	"errors"

	"github.com/ertan/farcaster-go/pkg/account"
	"github.com/ertan/farcaster-go/pkg/registry"
	"github.com/ertan/farcaster-go/pkg/users"
)

type AssetService struct {
	account  *account.AccountService
	registry *registry.RegistryService
}

type Collection struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ItemCount    int64  `json:"item_count"`
	OwnerCount   int64  `json:"owner_count"`
	FCOwnerCount int64  `json:"fc_owner_count"`
	ImageUrl     string `json:"image_url"`
	VolumeTraded string `json:"volume_traded"`
	ExternalUrl  string `json:"external_url,omitempty"`
	OpenSeaUrl   string `json:"opensea_url,omitempty"`
	TwitterUrl   string `json:"twitter_url,omitempty"`
	SchemaName   string `json:"schema_name,omitempty"`
}

func NewAssetService(account *account.AccountService, registry *registry.RegistryService) *AssetService {
	return &AssetService{
		account:  account,
		registry: registry,
	}
}

func (c *AssetService) GetCollectionOwners(collectionId string, limit int, cursor string) ([]users.User, string, error) {
	type CollectionOwnerResponse struct {
		Result struct {
			Users []users.User `json:"users"`
		} `json:"result"`
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	params := map[string]interface{}{
		"collectionId": collectionId,
	}
	if limit > 0 {
		params["limit"] = limit
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	var collectionOwnerResponse CollectionOwnerResponse
	responseBytes, err := c.account.SendRequest("GET", "/v2/collection-owners", params, nil)
	if err != nil {
		return nil, "", err
	}
	if err := json.Unmarshal(responseBytes, &collectionOwnerResponse); err == nil {
		if len(collectionOwnerResponse.Errors) > 0 {
			return nil, "", errors.New(collectionOwnerResponse.Errors[0].Message)
		}
		return collectionOwnerResponse.Result.Users, collectionOwnerResponse.Next.Cursor, nil
	}
	return nil, "", err
}

func (c *AssetService) GetCollectionsByOwnerFid(fid uint64, limit int, cursor string) ([]Collection, string, error) {
	params := map[string]interface{}{
		"ownerFid": fid,
	}
	if limit > 0 {
		params["limit"] = limit
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	var collectionsResponse struct {
		Result struct {
			Collections []Collection `json:"collections"`
		} `json:"result"`
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	responseBytes, err := c.account.SendRequest("GET", "/v2/user-collections", params, nil)
	if err != nil {
		return nil, "", err
	}
	if err := json.Unmarshal(responseBytes, &collectionsResponse); err == nil {
		if len(collectionsResponse.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", err
		}
		return collectionsResponse.Result.Collections, collectionsResponse.Next.Cursor, nil
	}
	return nil, "", err
}

func (c *AssetService) GetCollectionsByOwnerFname(fname string, limit int, cursor string) ([]Collection, string, error) {
	if c.registry == nil {
		return nil, "", errors.New("registry service is not initialized. Use GetCollectionsByOwnerFid instead")
	}
	fid, err := c.registry.GetFidByFname(fname)
	if err != nil {
		return nil, "", err
	}
	return c.GetCollectionsByOwnerFid(fid, limit, cursor)
}

func (c *AssetService) GetCollectionsByOwnerAddress(address string, limit int, cursor string) ([]Collection, string, error) {
	if c.registry == nil {
		return nil, "", errors.New("registry service is not initialized. Use GetCollectionsByOwnerFid instead")
	}
	fid, err := c.registry.GetFidByAddress(address)
	if err != nil {
		return nil, "", err
	}
	return c.GetCollectionsByOwnerFid(fid, 0, "")
}
