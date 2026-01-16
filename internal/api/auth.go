package api

import (
	"context"
	"time"
)

type APIKeyResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"`
	KeyPrefix string `json:"key_prefix"`
	Scopes    string `json:"scopes"`
	CreatedAt string `json:"created_at"`
}

type APIKeyListResponse struct {
	APIKeys []APIKeyDTO `json:"api_keys"`
}

type APIKeyDTO struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Scopes     string     `json:"scopes"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateAPIKeyRequest struct {
	Name   string `json:"name"`
	Scopes string `json:"scopes,omitempty"`
}

func (c *Client) CreateAPIKey(ctx context.Context, name, scopes string) (*APIKeyResponse, error) {
	req := CreateAPIKeyRequest{
		Name:   name,
		Scopes: scopes,
	}

	var resp APIKeyResponse
	if err := c.Post(ctx, "/api/v1/auth/api-keys", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ListAPIKeys(ctx context.Context) (*APIKeyListResponse, error) {
	var resp APIKeyListResponse
	if err := c.Get(ctx, "/api/v1/auth/api-keys", &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) RevokeAPIKey(ctx context.Context, id string) error {
	return c.Delete(ctx, "/api/v1/auth/api-keys/" + id)
}

type AuthStatusResponse struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"user_id,omitempty"`
	Email         string `json:"email,omitempty"`
}

func (c *Client) GetAuthStatus(ctx context.Context) (*AuthStatusResponse, error) {
	var resp struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}

	if err := c.Get(ctx, "/api/v1/auth/me", &resp); err != nil {
		return &AuthStatusResponse{Authenticated: false}, nil
	}

	return &AuthStatusResponse{
		Authenticated: true,
		UserID:        resp.ID,
		Email:         resp.Email,
	}, nil
}
