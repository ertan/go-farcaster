package health

import "github.com/ertan/go-farcaster/pkg/account"

type HealthService struct {
	account *account.AccountService
}

func NewHealthService(account *account.AccountService) *HealthService {
	return &HealthService{
		account: account,
	}
}

func (h *HealthService) OK() error {
	_, err := h.account.SendRequest("GET", "/v2/health", nil, nil)
	if err != nil {
		return err
	}
	return nil
}
