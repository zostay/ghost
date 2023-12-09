package keeper

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/memory"
)

var ErrDuplicate = errors.New("duplicate secret")

type secretKey struct {
	name     string
	username string
	location string
}

type localKey struct {
	id           string
	lastModified time.Time
}

func makeKey(sec secrets.Secret) secretKey {
	return secretKey{
		name:     sec.Name(),
		username: sec.Username(),
		location: sec.Location(),
	}
}

// Sync is an engine that helps with the copying of secrets between secret
// keepers. It organizes these copies on the basis of name, username, and
// location as the key values.
//
// This works by using calls to one or more of the Add* methods to configure the
// secrets to sync. Then CopyTo can be used to send these secrets to another
// secret keeper. The DeleteAbsent will delete any secrets in the given secret
// keeper that have not be added using the Add* methods.
type Sync struct {
	gatherer secrets.Keeper
	index    map[secretKey]localKey
}

// NewSync creates a new blank object for handling sync between secret keepers.
func NewSync() (*Sync, error) {
	mem, err := memory.New()
	if err != nil {
		return nil, err
	}

	return &Sync{
		gatherer: mem,
		index:    map[secretKey]localKey{},
	}, nil
}

func (s *Sync) addToIndex(
	ctx context.Context,
	sec secrets.Secret,
) error {
	var memSec secrets.Secret
	memSec = secrets.NewSingleFromSecret(sec, secrets.WithID(""))
	memSec, err := s.gatherer.SetSecret(ctx, memSec)
	if err != nil {
		return err
	}

	s.index[makeKey(sec)] = localKey{
		id:           memSec.ID(),
		lastModified: sec.LastModified(),
	}

	return nil
}

// AddSecret adds a single secret to the list to be copied. If the secret has
// already been added, it will return ErrDuplicate unless ignoreDuplicate is set
// to true. If ignoreDuplicate is set, the more recent secret will be kept.
func (s *Sync) AddSecret(
	ctx context.Context,
	sec secrets.Secret,
	ignoreDuplicate bool,
) error {
	sk := makeKey(sec)
	if similar, similarExists := s.index[sk]; !similarExists {
		if ignoreDuplicate {
			if sec.LastModified().After(similar.lastModified) {
				return s.addToIndex(ctx, sec)
			}
			return nil
		}

		return ErrDuplicate
	}

	return s.addToIndex(ctx, sec)
}

// AddLocationSecret adds all the secrets in a given location to the list to be
// copied. If the location contains secrets with identical name and username,
// ErrDuplicate will be returned unless ignoreDuplicates is set to true. If
// ignoreDuplicates is set, the most recent secret will be kept.
func (s *Sync) AddLocationSecret(
	ctx context.Context,
	from secrets.Keeper,
	loc string,
	ignoreDuplicates bool,
) error {
	ids, err := from.ListSecrets(ctx, loc)
	if err != nil {
		return err
	}

	for _, id := range ids {
		sec, err := from.GetSecret(ctx, id)
		if err != nil {
			return err
		}

		if err := s.AddSecret(ctx, sec, ignoreDuplicates); err != nil {
			return err
		}
	}

	return nil
}

// AddSecretKeeper adds all secrets in a keeper to the destination.
//
// If the secret keeper contains more than one secret with the same name,
// username, and location, the ErrDuplicate will be returned, with the Sync
// object now partially filled. You can set ignoreDuplicates to cause secondary
// secrets to be ignored. If set, the most recently modified secret will be
// kept.
func (s *Sync) AddSecretKeeper(
	ctx context.Context,
	from secrets.Keeper,
	ignoreDuplicates bool,
) error {
	fromLocs, err := from.ListLocations(ctx)
	if err != nil {
		return err
	}

	for _, loc := range fromLocs {
		if err := s.AddLocationSecret(ctx, from, loc, ignoreDuplicates); err != nil {
			return err
		}
	}

	return nil
}

// CopyTo copies all the secrets that have been added to the Sync object for
// copying via the Add* methods into the given keeper.
func (s *Sync) CopyTo(
	ctx context.Context,
	to secrets.Keeper,
	logger *log.Logger,
) error {
	for sk, lk := range s.index {
		secs, err := to.GetSecretsByName(ctx, sk.name)
		if err != nil {
			return err
		}

		var syncSec secrets.Secret
		for _, sec := range secs {
			if sec.Username() == sk.username && sec.Location() == sk.location {
				syncSec = sec
				break
			}
		}

		if syncSec == nil {
			syncSec = secrets.NewSecret("", "", "")
		}

		origSec, err := s.gatherer.GetSecret(ctx, lk.id)
		if err != nil {
			return err
		}

		if logger != nil {
			logger.Printf("Copying %s/%s/%s", sk.location, sk.name, sk.username)
		}

		secrets.SetName(syncSec, origSec.Name())
		secrets.SetUsername(syncSec, origSec.Username())
		secrets.SetPassword(syncSec, origSec.Password())
		secrets.SetType(syncSec, origSec.Type())
		secrets.SetUrl(syncSec, origSec.Url())
		for fldName, fldVal := range origSec.Fields() {
			// Problem: a field that has been deleted, might remain set in the
			// destination, but since I don't really have a facility for
			// deleting fields... that's just the way it is for now
			secrets.SetField(syncSec, fldName, fldVal)
		}

		if _, err := to.SetSecret(ctx, syncSec); err != nil {
			return err
		}
	}

	return nil
}

// DeleteAbsent deletes all the secrets in the destination keeper that do not
// exactly match the ones added to the Sync object via the Add* methods. It
// matches using name, username, and location.
func (s *Sync) DeleteAbsent(
	ctx context.Context,
	to secrets.Keeper,
	logger *log.Logger,
) error {
	locs, err := to.ListLocations(ctx)
	if err != nil {
		return err
	}

	for _, loc := range locs {
		ids, err := to.ListSecrets(ctx, loc)
		if err != nil {
			return err
		}

		for _, id := range ids {
			sec, err := to.GetSecret(ctx, id)
			if err != nil {
				return err
			}

			if _, secExists := s.index[makeKey(sec)]; !secExists {
				if logger != nil {
					logger.Printf("Deleting %s/%s/%s", sec.Location(), sec.Name(), sec.Username())
				}

				if err := to.DeleteSecret(ctx, id); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
