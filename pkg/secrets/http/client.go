package http

import (
	"context"
	"errors"
	"io"

	"github.com/zostay/ghost/pkg/secrets"
)

// Client is a secret keeper that communicates with the secret keeper service
// provided in Server.
type Client struct {
	client KeeperClient
}

var _ secrets.Keeper = &Client{}

// NewClient creates a new client for the secret keeper service.
func NewClient(client KeeperClient) *Client {
	return &Client{client}
}

// ListLocations retrieves the list of locations from the secret keeper service.
func (c *Client) ListLocations(ctx context.Context) ([]string, error) {
	locStream, err := c.client.ListLocations(ctx, nil)
	if err != nil {
		return nil, err
	}

	locations := []string{}
	for {
		loc, err := locStream.Recv()
		if errors.Is(err, io.EOF) {
			return locations, nil
		}
		if err != nil {
			return nil, err
		}

		locations = append(locations, loc.GetLocation())
	}
}

// ListSecrets retrieves the list of secrets from the secret keeper service.
func (c *Client) ListSecrets(ctx context.Context, location string) ([]string, error) {
	secStream, err := c.client.ListSecrets(ctx, &Location{
		Location: location,
	})
	if err != nil {
		return nil, err
	}

	secs := []string{}
	for {
		sec, err := secStream.Recv()
		if errors.Is(err, io.EOF) {
			return secs, nil
		}
		if err != nil {
			return nil, err
		}

		secs = append(secs, sec.GetId())
	}
}

// GetSecret retrieves the secret with the given ID from the secret keeper
// service.
func (c *Client) GetSecret(ctx context.Context, id string) (secrets.Secret, error) {
	sec, err := c.client.GetSecret(ctx, &GetSecretRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return NewSecretWrapper(sec), nil
}

// GetSecretsByName retrieves the list of secrets with the given name from the
// secret keeper service.
func (c *Client) GetSecretsByName(ctx context.Context, name string) ([]secrets.Secret, error) {
	rawSecs, err := c.client.GetSecretsByName(ctx, &GetSecretsByNameRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	secs := []secrets.Secret{}
	for {
		sec, err := rawSecs.Recv()
		if errors.Is(err, io.EOF) {
			return secs, nil
		}
		if err != nil {
			return nil, err
		}

		secs = append(secs, NewSecretWrapper(sec))
	}
}

// SetSecret stores the given secret in the secret keeper service.
func (c *Client) SetSecret(ctx context.Context, secret secrets.Secret) (secrets.Secret, error) {
	sec, err := c.client.SetSecret(ctx, FromSecret(secret))
	if err != nil {
		return nil, err
	}

	return NewSecretWrapper(sec), nil
}

// CopySecret copies the secret with the given ID to the given location in the
// secret keeper service.
func (c *Client) CopySecret(ctx context.Context, id, location string) (secrets.Secret, error) {
	sec, err := c.client.CopySecret(ctx, &ChangeLocationRequest{
		Id:       id,
		Location: location,
	})
	if err != nil {
		return nil, err
	}

	return NewSecretWrapper(sec), nil
}

// MoveSecret moves the secret with the given ID to the given location in the
// secret keeper service.
func (c *Client) MoveSecret(ctx context.Context, id, location string) (secrets.Secret, error) {
	sec, err := c.client.MoveSecret(ctx, &ChangeLocationRequest{
		Id:       id,
		Location: location,
	})
	if err != nil {
		return nil, err
	}

	return NewSecretWrapper(sec), nil
}

// DeleteSecret deletes the secret with the given ID from the secret keeper
// service.
func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	_, err := c.client.DeleteSecret(ctx, &DeleteSecretRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return nil
}
