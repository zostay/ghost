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
// keeper that have not been added using the Add* methods.
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

type syncOptions struct {
	ignoreDuplicates  bool
	logger            *log.Logger
	overwriteMatching bool
}

type SyncOption func(*syncOptions)

// WithIgnoredDuplicates causes the AddSecret* method to ignore duplicate secrets
// that have already been added. If set, the most recent secret will be kept.
func WithIgnoredDuplicates() SyncOption {
	return func(o *syncOptions) {
		o.ignoreDuplicates = true
	}
}

func processSyncOptions(opts []SyncOption) *syncOptions {
	o := &syncOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// AddSecret adds a single secret to the list to be copied. If the secret has
// already been added, it will return ErrDuplicate unless WithIgnoredDuplicate is set
// to true. If WithIgnoredDuplicates is set, the most recent secret will be kept.
//
// Valid options for this method include WithIgnoredDuplicates.
func (s *Sync) AddSecret(
	ctx context.Context,
	sec secrets.Secret,
	opts ...SyncOption,
) error {
	o := processSyncOptions(opts)
	if o.logger != nil {
		o.logger.Printf("Preparing to sync %s/%s/%s", sec.Location(), sec.Name(), sec.Username())
	}

	sk := makeKey(sec)
	if similar, similarExists := s.index[sk]; !similarExists {
		if o.ignoreDuplicates {
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
// ErrDuplicate will be returned unless WithIgnoredDuplicates is set to true. If
// WithIgnoredDuplicates is set, the most recent secret will be kept.
//
// Valid options for this method include WithIgnoredDuplicates.
func (s *Sync) AddLocationSecret(
	ctx context.Context,
	from secrets.Keeper,
	loc string,
	opts ...SyncOption,
) error {
	o := processSyncOptions(opts)
	if o.logger != nil {
		o.logger.Printf("Preparing to sync location %s", loc)
	}

	ids, err := from.ListSecrets(ctx, loc)
	if err != nil {
		return err
	}

	for _, id := range ids {
		sec, err := from.GetSecret(ctx, id)
		if err != nil {
			return err
		}

		if err := s.AddSecret(ctx, sec, opts...); err != nil {
			return err
		}
	}

	return nil
}

// AddSecretKeeper adds all secrets in a keeper to the destination.
//
// If the secret keeper contains more than one secret with the same name,
// username, and location, the ErrDuplicate will be returned, with the Sync
// object now partially filled. You can set WithIgnoredDuplicates to cause secondary
// secrets to be ignored. If set, the most recently modified secret will be
// kept.
//
// Valid options for this method include WithIgnoredDuplicates.
func (s *Sync) AddSecretKeeper(
	ctx context.Context,
	from secrets.Keeper,
	opts ...SyncOption,
) error {
	fromLocs, err := from.ListLocations(ctx)
	if err != nil {
		return err
	}

	for _, loc := range fromLocs {
		if err := s.AddLocationSecret(ctx, from, loc, opts...); err != nil {
			return err
		}
	}

	return nil
}

// WithLogger sets the logger to use when copying secrets.
func WithLogger(logger *log.Logger) SyncOption {
	return func(o *syncOptions) {
		o.logger = logger
	}
}

// WithMatchingOverwritten causes the CopyTo method to overwrite existing secrets in the
// destination keeper. The secrets will be overwritten, if they have the same
// name, username, and location in the destination. If there are multiple secrets
// with the same name, username, and location in the destination, the most
// recently modified secret will be overwritten.
func WithMatchingOverwritten() SyncOption {
	return func(o *syncOptions) {
		o.overwriteMatching = true
	}
}

// CopyTo copies all the secrets that have been added to the Sync object for
// copying via the Add* methods into the given keeper. If a logger is given,
// this will write a message to that logger each time a secret is copied. If the
// secret already exists in the destination, it will not be overwritten unless
// the WithMatchingOverwritten option is set.
//
// Valid options for this method include WithLogger and WithMatchingOverwritten.
func (s *Sync) CopyTo(
	ctx context.Context,
	to secrets.Keeper,
	opts ...SyncOption,
) error {
	o := processSyncOptions(opts)
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

		action := "Copying"

		// detect whether we are overwriting
		var overSecs []secrets.Secret
		if o.overwriteMatching {
			overSecs, err = to.GetSecretsByName(ctx, sk.name)
			if err != nil {
				return err
			}

			if len(overSecs) > 0 {
				action = "Overwriting"
			}
		}

		if o.logger != nil {
			o.logger.Printf("%s %s/%s/%s", action, sk.location, sk.name, sk.username)
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

		// select the secret to overwrite when overwriting
		if o.overwriteMatching {
			var best secrets.Secret
			for _, sec := range overSecs {
				if sec.Username() == sk.username && sec.Location() == sk.location {
					if best == nil || sec.LastModified().After(best.LastModified()) {
						best = sec
					}
				}
			}

			if best != nil {
				syncSec = secrets.NewSingleFromSecret(best, secrets.WithID(best.ID()))
			}
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
//
// If a logger is given, this will write a message to that logger each time a
// secret is deleted.
//
// Valid options for this method include WithLogger.
func (s *Sync) DeleteAbsent(
	ctx context.Context,
	to secrets.Keeper,
	opts ...SyncOption,
) error {
	o := processSyncOptions(opts)
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
				if o.logger != nil {
					o.logger.Printf("Deleting %s/%s/%s", sec.Location(), sec.Name(), sec.Username())
				}

				if err := to.DeleteSecret(ctx, id); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
