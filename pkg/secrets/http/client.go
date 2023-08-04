package http

import (
	"context"
	"errors"
	"io"

	"github.com/zostay/ghost/pkg/secrets"
)

type Client struct {
	client KeeperClient
}

var _ secrets.Keeper = &Client{}

func NewClient() *Client {
	return &Client{}
}

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

func (c *Client) GetSecret(ctx context.Context, id string) (secrets.Secret, error) {
	sec, err := c.client.GetSecret(ctx, &GetSecretRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return NewSecretWrapper(sec), nil
}

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

func (c *Client) SetSecret(ctx context.Context, secret secrets.Secret) (secrets.Secret, error) {
	sec, err := c.client.SetSecret(ctx, FromSecret(secret))
	if err != nil {
		return nil, err
	}

	return NewSecretWrapper(sec), nil
}

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

func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	_, err := c.client.DeleteSecret(ctx, &DeleteSecretRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	return nil
}
