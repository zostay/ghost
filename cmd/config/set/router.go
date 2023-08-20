package set

import (
	"errors"
	"sort"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/set"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/router"
)

var (
	RouterCmd = &cobra.Command{
		Use:     "router <keeper-name> [flags]",
		Short:   "Configure a router secret keeper",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: PreRunSetRouterKeeperConfig,
		Run:     RunSetKeeperConfig,
	}

	removeLocations []string
	addLocations    []string
	addKeeper       string
	defaultKeeper   string
)

func init() {
	RouterCmd.Flags().StringSliceVar(&removeLocations, "remove", []string{}, "Remove one or more locations from the router")
	RouterCmd.Flags().StringSliceVar(&addLocations, "add", []string{}, "Add one or more locations to the router")
	RouterCmd.Flags().StringVar(&addKeeper, "keeper", "", "Keeper to use with to the added locations")
	RouterCmd.Flags().StringVar(&defaultKeeper, "default", "", "Default keeper to use with the router")
}

func PreRunSetRouterKeeperConfig(cmd *cobra.Command, args []string) error {
	if len(removeLocations) > 0 && len(addLocations) > 0 {
		return errors.New("cannot remove and add locations in the same step")
	}

	if len(removeLocations) > 0 && defaultKeeper != "" {
		return errors.New("cannot remove locations and set the default keeper in the same step")
	}

	if len(addLocations) > 0 && addKeeper == "" {
		return errors.New("must specify a keeper to use with the added locations")
	}

	keeperName := args[0]
	c := config.Instance()
	kc := c.Keepers[keeperName]
	if kc == nil {
		kc = map[string]any{
			"type":   router.ConfigType,
			"routes": []map[string]any{},
		}
		Replacement = kc
	}

	rc := RouterConfig(kc)

	if defaultKeeper != "" {
		kc["default"] = defaultKeeper
	}

	if len(removeLocations) > 0 {
		rc.Remove(removeLocations...)
		return nil
	}

	rc.Add(addKeeper, addLocations...)

	Replacement = kc
	return nil
}

type RouterConfig config.KeeperConfig

func (rc RouterConfig) Add(keeper string, locations ...string) {
	routes := rc["routes"].([]map[string]any)
	if routes == nil {
		routes = make([]map[string]any, 0, 1)
	}

	routes = append(routes, map[string]any{
		"locations": locations,
		"keeper":    keeper,
	})

	rc["routes"] = routes
}

func (rc RouterConfig) Remove(removeLocations ...string) {
	routes := rc["routes"].([]map[string]any)
	removeSet := set.New(removeLocations...)

	deleteRoutes := []int{}
	for i, r := range routes {
		locations := r["locations"].([]string)
		for _, loc := range locations {
			if removeSet.Contains(loc) {
				locations = slices.DeleteValue(locations, loc)
				if len(locations) == 0 {
					deleteRoutes = append(deleteRoutes, i)
				}
				routes[i]["locations"] = locations
			}
		}
	}

	sort.Reverse(sort.IntSlice(deleteRoutes))

	for _, i := range deleteRoutes {
		routes = slices.Delete(routes, i)
	}

	rc["routes"] = routes
}
