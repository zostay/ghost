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

	username, password     string
	prompt                 bool
	location               string
	typ                    string
	moveSecret, copySecret bool
	url                    string
	setFlds                map[string]string
)

func init() {
	setCmd.Flags().StringVar(&keeperName, "keeper", "", "The name of the secret keeper to use")
	setCmd.Flags().StringVar(&id, "id", "", "The ID of an existing secret to update")
	setCmd.Flags().StringVar(&name, "name", "", "The new name to set")
	setCmd.Flags().StringVar(&username, "username", "", "The new username to set")
	setCmd.Flags().StringVar(&password, "password", "", "The new password to set")
	setCmd.Flags().BoolVar(&prompt, "prompt", false, "Prompt for the password")
	setCmd.Flags().StringVar(&typ, "type", "", "The new type of secret to set")
	setCmd.Flags().StringVar(&location, "location", "", "The location to give the secret")
	setCmd.Flags().BoolVar(&moveSecret, "move", false, "Move the secret to a new location")
	setCmd.Flags().BoolVar(&copySecret, "copy", false, "Copy the secret to a new location")
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

	if moveSecret && copySecret {
		s.Logger.Panic("Cannot specify both --move and --copy.")
	}

	var opVerb string
	switch {
	case moveSecret:
		opVerb = "moving"
	case copySecret:
		opVerb = "copying"
	default:
		opVerb = "saving"
	}

	if (moveSecret || copySecret) && location == "" {
		s.Logger.Panicf("You must specify a --location to place the secret while %s.", opVerb)
	}

	if password != "" && prompt {
		s.Logger.Panic("Cannot specify both --password and --prompt.")
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

	ctx := keeper.WithBuilder(context.Background(), c)
	kpr, err := keeper.Build(ctx, keeperName)
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
		if moveSecret || copySecret {
			s.Logger.Panicf("The secret is new. No %s allowed.", opVerb)
		}
		sec = secrets.NewSecret(name, "", "", secrets.WithLocation(location))
	case 1:
		sec = secs[0]
		if location != "" && location != sec.Location() && !moveSecret && !copySecret {
			s.Logger.Panicf("Cannot change location when %s. You must specify --move or --copy.", opVerb)
		}
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
		pw, err := keeper.GetPassword(sec.Name(), "Password", "Please enter the new password", "Set Password")
		if err != nil {
			s.Logger.Panicf("Unable to prompt for password: %v", err)
		}
		sec = secrets.SetPassword(sec, pw)
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

	if moveSecret {
		newSec, err = kpr.MoveSecret(ctx, newSec.ID(), location)
		if err != nil {
			s.Logger.Panic(err)
		}
	}

	if copySecret {
		newSec, err = kpr.CopySecret(ctx, newSec.ID(), location)
		if err != nil {
			s.Logger.Panic(err)
		}
	}

	s.PrintSecret(newSec, false)
}
