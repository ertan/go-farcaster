package notifications

import (
	"encoding/json"
	"errors"

	"github.com/ertan/go-farcaster/pkg/account"
	"github.com/ertan/go-farcaster/pkg/casts"
	"github.com/ertan/go-farcaster/pkg/users"
)

type NotificationService struct {
	account *account.AccountService
}

type Notification struct {
	Type      string      `json:"type"`
	Id        string      `json:"id"`
	Timestamp uint64      `json:"timestamp"`
	Actor     *users.User `json:"actor"`
	Content   struct {
		Cast *casts.Cast `json:"cast"`
	} `json:"content"`
	Next struct {
		Cursor string `json:"cursor"`
	} `json:"next"`
}

func NewNotificationService(account *account.AccountService) *NotificationService {
	return &NotificationService{
		account: account,
	}
}

func (n *NotificationService) GetNotifications(limit int, cursor string) ([]Notification, string, error) {
	type NotificationsResponse struct {
		Result struct {
			Notifications []Notification `json:"notifications"`
		} `json:"result"`
		Next struct {
			Cursor string `json:"cursor"`
		} `json:"next"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	params := map[string]interface{}{}
	if limit > 0 {
		params["limit"] = limit
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	responseBytes, err := n.account.SendRequest("GET", "/v2/notifications", params, nil)
	if err != nil {
		return nil, "", err
	}
	var response NotificationsResponse
	if err := json.Unmarshal(responseBytes, &response); err == nil {
		if len(response.Errors) > 0 {
			// TODO(ertan): Find a better solution to pass the errors here.
			return nil, "", errors.New(response.Errors[0].Message)
		}
		return response.Result.Notifications, response.Next.Cursor, nil
	}
	return nil, "", errors.New("Error getting notifications")
}
