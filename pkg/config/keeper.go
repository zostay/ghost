package config

import (
	"fmt"

	"github.com/zostay/go-std/maps"
	"github.com/zostay/go-std/set"
	"github.com/zostay/go-std/slices"
)

type KeeperType int

const (
	KTNone KeeperType = iota
	KTConflict
	KTLastPass
	KTKeepass
	KTLowSecurity
	KTGRPC
	KTKeyring
	KTMemory
	KTHuman
	KTRouter
	KTSeq
)

func (kt KeeperType) String() string {
	switch kt {
	case KTNone:
		return "none"
	case KTConflict:
		return "conflict"
	case KTLastPass:
		return "lastpass"
	case KTKeepass:
		return "keepass"
	case KTLowSecurity:
		return "low"
	case KTGRPC:
		return "grpc"
	case KTKeyring:
		return "keyring"
	case KTMemory:
		return "memory"
	case KTHuman:
		return "human"
	case KTRouter:
		return "router"
	case KTSeq:
		return "seq"
	default:
		return "unknown"
	}
}

var KeeperTypes = []KeeperType{
	KTLastPass,
	KTKeepass,
	KTLowSecurity,
	KTGRPC,
	KTKeyring,
	KTMemory,
	KTHuman,
	KTRouter,
	KTSeq,
}

type KeeperConfig struct {
	LastPass LastPassConfig    `yaml:"lastpass,omitempty"`
	Keepass  KeepassConfig     `yaml:"keepass,omitempty"`
	Low      LowSecurityConfig `yaml:"low,omitempty"`
	GRPC     GRPCConfig        `yaml:"grpc,omitempty"`
	Keyring  KeyringConfig     `yaml:"keyring,omitempty"`
	Memory   InternalConfig    `yaml:"memory,omitempty"`
	Human    HumanConfig       `yaml:"human,omitempty"`
	Router   RouterConfig      `yaml:"router,omitempty"`
	Seq      SeqConfig         `yaml:"seq,omitempty"`
}

func (kc *KeeperConfig) Check(c *Config) error {
	errs := NewValidationError()

	switch kc.Type() {
	case KTNone:
		errs.Append(fmt.Errorf("no keeper type configured"))

	case KTConflict:
		errs.Append(fmt.Errorf("cannot configure more than one keeper type"))

	case KTLastPass:
		// no additional validation rules...

	case KTKeepass:
		// no additional validation rules...

	case KTLowSecurity:
		// no additional validation rules...

	case KTGRPC:
		if kc.GRPC.Listener != "unix" {
			errs.Append(fmt.Errorf("grpc listener type %q is not supported", kc.GRPC.Listener))
		}

	case KTKeyring:
		// no additional validation rules...

	case KTMemory:
		// no additional validation rules...

	case KTHuman:
		for _, q := range kc.Human.Questions {
			flds := set.New[string](maps.Keys(q.Presets)...)
			for _, f := range q.AskFor {
				if !flds.Contains(f) {
					errs.Append(fmt.Errorf("human question configuration already contains field named %q", f))
				}
				flds.Insert(f)
			}
		}

	case KTRouter:
		if kc.Router.DefaultRoute != "" {
			if _, keeperExists := c.Keepers[kc.Router.DefaultRoute]; !keeperExists {
				errs.Append(fmt.Errorf("default route keeper %q does not exist", kc.Router.DefaultRoute))
			}
		}

		for _, r := range kc.Router.Routes {
			if _, keeperExists := c.Keepers[r.Keeper]; !keeperExists {
				errs.Append(fmt.Errorf("route keeper %q does not exist", r.Keeper))
			}

			if len(r.Locations) == 0 {
				errs.Append(fmt.Errorf("route keeper %q has no locations", r.Keeper))
			}
		}

	case KTSeq:
		for _, k := range kc.Seq.Keepers {
			if _, keeperExists := c.Keepers[k]; !keeperExists {
				errs.Append(fmt.Errorf("seq keeper %q does not exist", k))
			}
		}
	}

	return errs.Return()
}

func (kc *KeeperConfig) Type() KeeperType {
	t := KTNone
	if kc.LastPass.Username != "" {
		if t != KTNone {
			return KTConflict
		}
		t = KTLastPass
	}

	if kc.Keepass.Path != "" {
		if t != KTNone {
			return KTConflict
		}
		t = KTKeepass
	}

	if kc.Low.Path != "" {
		if t != KTNone {
			return KTConflict
		}
		t = KTLowSecurity
	}

	if kc.GRPC.Listener != "" {
		if t != KTNone {
			return KTConflict
		}
		t = KTGRPC
	}

	if kc.Keyring.ServiceName != "" {
		if t != KTNone {
			return KTConflict
		}
		t = KTKeyring
	}

	if kc.Memory.Enable {
		if t != KTNone {
			return KTConflict
		}
		t = KTMemory
	}

	if len(kc.Human.Questions) > 0 {
		if t != KTNone {
			return KTConflict
		}
		t = KTHuman
	}

	if len(kc.Router.Routes) > 0 || kc.Router.DefaultRoute != "" {
		if t != KTNone {
			return KTConflict
		}
		t = KTRouter
	}

	if len(kc.Seq.Keepers) > 0 {
		if t != KTNone {
			return KTConflict
		}
		t = KTSeq
	}

	return t
}

type LastPassConfig struct {
	Username string `yaml:"username"`
}

type KeepassConfig struct {
	Path string `yaml:"path"`
}

type LowSecurityConfig struct {
	Path string `yaml:"path"`
}

type GRPCConfig struct {
	Listener string `yaml:"listener"`
}

type KeyringConfig struct {
	ServiceName string
}

type InternalConfig struct {
	Enable bool `yaml:"enable"`
}

type HumanQuestionConfig struct {
	ID      string            `yaml:"ID"`
	Presets map[string]string `yaml:"presets"`
	AskFor  []string          `yaml:"ask_for"`
}

type HumanConfig struct {
	Questions []HumanQuestionConfig `yaml:"questions"`
}

func (hc *HumanConfig) Set(
	id string,
	presets map[string]string,
	askFor []string,
) {
	if hc.Questions == nil {
		hc.Questions = []HumanQuestionConfig{}
	}

	hc.Questions = append(hc.Questions, HumanQuestionConfig{
		ID:      id,
		Presets: presets,
		AskFor:  askFor,
	})
}

func (hc *HumanConfig) Remove(id string) {
	if hc.Questions == nil {
		return
	}

	i := slices.FirstIndex(hc.Questions, func(q HumanQuestionConfig) bool {
		return q.ID == id
	})
	if i >= 0 {
		hc.Questions = slices.Delete(hc.Questions, i)
	}
}

type RouterConfig struct {
	Routes       []RouteConfig `yaml:"routes"`
	DefaultRoute string        `yaml:"default,omitempty"`
}

func (rc *RouterConfig) Add(keeper string, locations ...string) {
	rc.Routes = append(rc.Routes, RouteConfig{
		Locations: locations,
		Keeper:    keeper,
	})
}

func (rc *RouterConfig) Remove(removeLocations ...string) {
	removeSet := set.New(removeLocations...)

	for i, r := range rc.Routes {
		for _, loc := range r.Locations {
			if removeSet.Contains(loc) {
				rc.Routes[i].Locations = slices.DeleteValue(
					rc.Routes[i].Locations, loc)
			}
		}
	}
}

type RouteConfig struct {
	Locations []string `yaml:"locations"`
	Keeper    string   `yaml:"keeper"`
}

type SeqConfig struct {
	Keepers []string `yaml:"keepers"`
}
