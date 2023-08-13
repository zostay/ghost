package secrets

import (
	"context"
	"errors"
)

var ErrSkipLocation = errors.New("skip location")

func ForEachInLocation(
	ctx context.Context,
	kpr Keeper,
	location string,
	run func(Secret) error,
) error {
	ids, err := kpr.ListSecrets(ctx, location)
	if err != nil {
		return err
	}

	for _, id := range ids {
		sec, err := kpr.GetSecret(ctx, id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}
			return err
		}

		if err := run(sec); err != nil {
			return err
		}
	}

	return nil
}

func ForEach(
	ctx context.Context,
	kpr Keeper,
	run func(Secret) error,
) error {
	locs, err := kpr.ListLocations(ctx)
	if err != nil {
		return err
	}

	for _, loc := range locs {
		if err := ForEachInLocation(ctx, kpr, loc, run); err != nil {
			if errors.Is(err, ErrSkipLocation) {
				continue
			}
			return err
		}
	}

	return nil
}
