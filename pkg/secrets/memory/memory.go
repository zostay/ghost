package memory

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"

	"github.com/oklog/ulid/v2"

	"github.com/zostay/ghost/pkg/secrets"
)

// encryption in here is probably just me being paranoid

// Memory is a Keeper that stores secrets in memory.
type Memory struct {
	cipher  cipher.AEAD
	nonce   []byte
	secrets map[string][]byte
}

var _ secrets.Keeper = &Memory{}

// MustNew calls New and panics if it returns an error.
func MustNew() *Memory {
	i, err := New()
	if err != nil {
		panic(err)
	}
	return i
}

// New constructs a new secret memory store.
func New() (*Memory, error) {
	k := make([]byte, 32)
	_, err := rand.Read(k)
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}

	gc, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gc.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	i := &Memory{
		cipher:  gc,
		nonce:   nonce,
		secrets: make(map[string][]byte),
	}

	return i, nil
}

func (i *Memory) decodeSecret(s []byte) (secrets.Secret, error) {
	ds, err := i.cipher.Open(nil, i.nonce, s, nil)
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewReader(ds))

	var sec secrets.Single
	err = dec.Decode(&sec)
	if err != nil {
		return nil, err
	}

	return &sec, nil
}

func (i *Memory) encodeSecret(sec secrets.Secret) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(sec)
	if err != nil {
		return nil, err
	}

	s := buf.Bytes()
	es := i.cipher.Seal(nil, i.nonce, s, nil)

	return es, nil
}

// ListLocations returns a list of all the secret names in the store.
func (i *Memory) ListLocations(context.Context) ([]string, error) {
	locs := make([]string, 0, len(i.secrets)>>1)
	for _, ct := range i.secrets {
		sec, err := i.decodeSecret(ct)
		if err != nil {
			return nil, err
		}

		locs = append(locs, sec.Location())
	}
	return locs, nil
}

// ListSecrets returns a list of all the secret IDs at the given location.
func (i *Memory) ListSecrets(_ context.Context, loc string) ([]string, error) {
	ids := make([]string, 0, len(i.secrets)>>1)
	for _, ct := range i.secrets {
		sec, err := i.decodeSecret(ct)
		if err != nil {
			return nil, err
		}

		if sec.Location() == loc {
			ids = append(ids, sec.ID())
		}
	}
	return ids, nil
}

// GetSecret retrieves the identified secret from the internal memory store.
func (i *Memory) GetSecret(_ context.Context, id string) (secrets.Secret, error) {
	if s, ok := i.secrets[id]; ok {
		return i.decodeSecret(s)
	}
	return nil, secrets.ErrNotFound
}

// GetSecretsByName retrieves all secrets with the given name.
func (i *Memory) GetSecretsByName(_ context.Context, name string) ([]secrets.Secret, error) {
	secs := make([]secrets.Secret, 0, 1)
	for _, ct := range i.secrets {
		sec, err := i.decodeSecret(ct)
		if err != nil {
			return nil, err
		}

		if sec.Name() == name {
			secs = append(secs, sec)
		}
	}
	return secs, nil
}

// SetSecret saves the named secret to the given value in the internal memory
// store.
func (i *Memory) SetSecret(_ context.Context, secret secrets.Secret) (secrets.Secret, error) {
	opts := make([]secrets.SingleOption, 0, 1)
	if _, hasSecret := i.secrets[secret.ID()]; secret.ID() == "" || !hasSecret {
		opts = append(opts, secrets.WithID(ulid.Make().String()))
	}
	single := secrets.NewSingleFromSecret(secret, opts...)

	es, err := i.encodeSecret(single)
	if err != nil {
		return nil, err
	}

	i.secrets[single.ID()] = es
	return single, nil
}

// CopySecret copies the secret into a new location while leaving the original
// in the old location.
func (i *Memory) CopySecret(ctx context.Context, id string, location string) (secrets.Secret, error) {
	secret, err := i.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	cp := secrets.NewSingleFromSecret(secret,
		secrets.WithLocation(location),
		secrets.WithID(ulid.Make().String()))

	es, err := i.encodeSecret(secret)
	if err != nil {
		return nil, err
	}

	i.secrets[cp.ID()] = es
	return cp, nil
}

// MoveSecret moves the secret into another location of the memory store.
func (i *Memory) MoveSecret(ctx context.Context, id string, location string) (secrets.Secret, error) {
	secret, err := i.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	mv := secrets.NewSingleFromSecret(secret,
		secrets.WithLocation(location))

	es, err := i.encodeSecret(secret)
	if err != nil {
		return nil, err
	}

	i.secrets[mv.ID()] = es
	return mv, nil
}

// DeleteSecret removes the identified secret from the store.
func (i *Memory) DeleteSecret(_ context.Context, id string) error {
	delete(i.secrets, id)
	return nil
}
