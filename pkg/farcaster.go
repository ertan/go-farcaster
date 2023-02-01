package farcaster

import (
	"github.com/ertan/go-farcaster/pkg/account"
	"github.com/ertan/go-farcaster/pkg/assets"
	"github.com/ertan/go-farcaster/pkg/casts"
	"github.com/ertan/go-farcaster/pkg/follows"
	"github.com/ertan/go-farcaster/pkg/health"
	"github.com/ertan/go-farcaster/pkg/notifications"
	"github.com/ertan/go-farcaster/pkg/reactions"
	"github.com/ertan/go-farcaster/pkg/registry"
	"github.com/ertan/go-farcaster/pkg/users"
	"github.com/ertan/go-farcaster/pkg/verifications"
)

type FarcasterClient struct {
	Account       *account.AccountService
	Assets        *assets.AssetService
	Casts         *casts.CastService
	Follows       *follows.FollowService
	Health        *health.HealthService
	Notifications *notifications.NotificationService
	Reactions     *reactions.ReactionService
	Registry      *registry.RegistryService
	Users         *users.UserService
	Verifications *verifications.VerificationsService
}

func NewFarcasterClient(apiUrl, mnemonic, providerWs string) *FarcasterClient {
	account := account.NewAccountService(apiUrl, mnemonic)
	registry := registry.NewRegistryService(providerWs)
	return &FarcasterClient{
		Account:       account,
		Assets:        assets.NewAssetService(account, registry),
		Casts:         casts.NewCastService(account, registry),
		Follows:       follows.NewFollowService(account, registry),
		Health:        health.NewHealthService(account),
		Notifications: notifications.NewNotificationService(account),
		Reactions:     reactions.NewReactionService(account),
		Registry:      registry,
		Users:         users.NewUserService(account, registry),
		Verifications: verifications.NewVerificationsService(account, registry),
	}
}
