package cmd

import (
	"context"
	neturl "net/url"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/secrets"
)

var (
	setCmd = &cobra.Command{
		Use:   "set",
		Short: "Set a secret",
		Run:   RunSet,
	}

	username, password         string
	prompt                     bool
	typ                        string
	moveLocation, copyLocation string
	url                        string
	setFlds                    map[string]string
)

func init() {
	setCmd.Flags().StringVar(&username, "username", "", "The new username to set")
	setCmd.Flags().StringVar(&password, "password", "", "The new password to set")
	setCmd.Flags().BoolVar(&prompt, "prompt", false, "Prompt for the password")
	setCmd.Flags().StringVar(&typ, "type", "", "The new type of secret to set")
	setCmd.Flags().StringVar(&moveLocation, "move", "", "Move the secret to a new location")
	setCmd.Flags().StringVar(&copyLocation, "copy", "", "Copy the secret to a new location")
	setCmd.Flags().StringVar(&url, "url", "", "The new URL to set")
	setCmd.Flags().StringToStringVar(&setFlds, "field", map[string]string{}, "The new fields to set")
}

func RunSet(cmd *cobra.Command, args []string) {
	if name != "" && id != "" {
		s.Logger.Panic("Cannot specify both --id and --name.")
	}

	if name == "" && id == "" {
		s.Logger.Panic("Must specify either --id or --name.")
	}

	if moveLocation != "" && copyLocation != "" {
		s.Logger.Panic("Cannot specify both --move and --copy.")
	}

	c := config.Instance()
	if keeperName == "" {
		keeperName = c.MasterKeeper
	}

	if keeperName == "" {
		s.Logger.Panic("No keeper specified.")
	}

	if _, hasConfig := c.Keepers[keeperName]; !hasConfig {
		s.Logger.Panicf("No keeper named %q.", keeperName)
	}

	ctx := context.Background()
	kpr, err := keeper.Build(ctx, keeperName, c)
	if err != nil {
		s.Logger.Panic(err)
	}

	secs := []secrets.Secret{}
	if id != "" {
		sec, err := kpr.GetSecret(ctx, id)
		if err != nil {
			s.Logger.Panic(err)
		}

		secs = append(secs, sec)
	} else {
		secs, err = kpr.GetSecretsByName(ctx, name)
		if err != nil {
			s.Logger.Panic(err)
		}
	}

	var sec secrets.Secret
	switch len(secs) {
	case 0:
		s.Logger.Panic("No matching secret found.")
	case 1:
		sec = secs[0]
	default:
		s.Logger.Panic("More than one matching secret found. Use --id to specify which one to update.")
	}

	if username != "" {
		sec = secrets.SetUsername(sec, username)
	}
	if password != "" {
		sec = secrets.SetPassword(sec, password)
	}
	if prompt {
		panic("not implemented")
	}
	if typ != "" {
		sec = secrets.SetType(sec, typ)
	}
	if url != "" {
		u, err := neturl.Parse(url)
		if err != nil {
			s.Logger.Panicf("Unable to parse URL %q: %v", url, err)
		}

		sec = secrets.SetUrl(sec, u)
	}
	for k, v := range setFlds {
		sec = secrets.SetField(sec, k, v)
	}

	newSec, err := kpr.SetSecret(ctx, sec)
	if err != nil {
		s.Logger.Panic(err)
	}

	if moveLocation != "" {
		newSec, err = kpr.MoveSecret(ctx, newSec.ID(), moveLocation)
		if err != nil {
			s.Logger.Panic(err)
		}
	}

	if copyLocation != "" {
		newSec, err = kpr.CopySecret(ctx, newSec.ID(), copyLocation)
		if err != nil {
			s.Logger.Panic(err)
		}
	}

	s.PrintSecret(newSec, false)
}
