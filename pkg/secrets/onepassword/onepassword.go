package onepassword

import (
	"context"

	"github.com/1Password/connect-sdk-go/connect"

	"github.com/zostay/ghost/pkg/secrets"
)

type OnePassword struct {
	pc connect.Client
}

func NewOnePassword(url string, token string) *OnePassword {
	return &OnePassword{
		pc: connect.NewClient(url, token),
	}
}

func (o *OnePassword) ListLocations(_ context.Context) ([]string, error) {
	vs, err := o.pc.GetVaults()
	if err != nil {
		return nil, err
	}

	locations := make([]string, 0, len(vs))
	for _, v := range vs {
		locations = append(locations, v.ID)
	}

	return locations, nil
}

func (o *OnePassword) ListSecrets(_ context.Context, location string) ([]string, error) {
	is, err := o.pc.GetItems(location)
	if err != nil {
		return nil, err
	}

	secs := make([]string, 0, len(is))
	for _, i := range is {
		secs = append(secs, i.ID)
	}

	return secs, nil
}

func (o *OnePassword) GetSecretsByName(_ context.Context, name string) ([]secrets.Secret, error) {
	vs, err := o.pc.GetVaults()
	if err != nil {
		return nil, err
	}

	var secs []secrets.Secret
	for _, v := range vs {
		is, err := o.pc.GetItemsByTitle(name, v.ID)
		if err != nil {
			return nil, err
		}

		for idx := range is {
			secs = append(secs, newSecret(&is[idx]))
		}
	}

	return secs, nil
}

func (o *OnePassword) GetSecret(_ context.Context, id string) (secrets.Secret, error) {
	vs, err := o.pc.GetVaults()
	if err != nil {
		return nil, err
	}

	for _, v := range vs {
		i, err := o.pc.GetItemByUUID(id, v.ID)
		if err != nil {
			continue
		}

		return newSecret(i), nil
	}

	return nil, secrets.ErrNotFound
}

func (o *OnePassword) SetSecret(_ context.Context, secret secrets.Secret) (secrets.Secret, error) {
	loc := secret.Location()
	if loc == "" {
		vs, err := o.pc.GetVaults()
		if err != nil {
			return nil, err
		}

		if len(vs) == 0 {
			return nil, secrets.ErrNotFound
		}

		loc = vs[0].ID
	}

	ns := fromSecret(secret)
	nns, err := o.pc.CreateItem(ns.item, loc)
	if err != nil {
		return nil, err
	}

	return newSecret(nns), nil
}

func (o *OnePassword) CopySecret(
	_ context.Context,
	id string,
	location string,
) (secrets.Secret, error) {
	vs, err := o.pc.GetVaults()
	if err != nil {
		return nil, err
	}

	for _, v := range vs {
		i, err := o.pc.GetItemByUUID(id, v.ID)
		if err != nil {
			continue
		}

		i.Vault.ID = location
		nns, err := o.pc.CreateItem(i, location)
		if err != nil {
			return nil, err
		}

		return newSecret(nns), nil
	}

	return nil, secrets.ErrNotFound
}

func (o *OnePassword) MoveSecret(_ context.Context, id string, location string) (secrets.Secret, error) {
	vs, err := o.pc.GetVaults()
	if err != nil {
		return nil, err
	}

	for _, v := range vs {
		i, err := o.pc.GetItemByUUID(id, v.ID)
		if err != nil {
			continue
		}

		i.Vault.ID = location
		nns, err := o.pc.CreateItem(i, location)
		if err != nil {
			return nil, err
		}

		err = o.pc.DeleteItemByID(id, v.ID)
		if err != nil {
			return nil, err
		}

		return newSecret(nns), nil
	}

	return nil, secrets.ErrNotFound
}

func (o *OnePassword) DeleteSecret(_ context.Context, id string) error {
	vs, err := o.pc.GetVaults()
	if err != nil {
		return err
	}

	for _, v := range vs {
		_, err := o.pc.GetItemByUUID(id, v.ID)
		if err != nil {
			continue
		}

		return o.pc.DeleteItemByID(id, v.ID)
	}

	return secrets.ErrNotFound
}

var _ secrets.Keeper = (*OnePassword)(nil)
