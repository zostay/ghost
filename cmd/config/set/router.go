package set

import (
	"errors"

	"github.com/spf13/cobra"
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

	if defaultKeeper != "" {
		Replacement.Router.DefaultRoute = defaultKeeper
	}

	if len(removeLocations) > 0 {
		Replacement.Router.Remove(removeLocations...)
		return nil
	}

	Replacement.Router.Add(addKeeper, addLocations...)
	return nil
}
