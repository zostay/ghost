package secrets

import (
	"context"
	"errors"
)

// ErrSkipLocation may be returned by a ForEach iterator function to skip the
// rest of the secrets in a location.
var ErrSkipLocation = errors.New("skip location")

// ForEachInLocation runs the given function for each secret in the named
// location.
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

// ForEach runs the given function for each secret in the keeper.
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
