package config

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zostay/ghost/cmd/flag"
	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
)

var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Add or update a secret keeper configuration",
}

var Replacement map[string]any

func init() {
	if err := setupCommands(); err != nil {
		panic(err)
	}
}

type LiteralOrSecretRef struct {
	Literal string
	Ref     config.SecretRef
}

func setupCommands() error {
	for _, pn := range plugin.List() {
		pc, _ := plugin.Get(pn)

		fields := make(map[string]*LiteralOrSecretRef, len(pc.CmdConfig.Fields))
		subCmd := &cobra.Command{
			Use:     pn + " <keeper-name>",
			Short:   pc.CmdConfig.Short,
			Args:    cobra.ExactArgs(1),
			Run:     RunSetKeeperConfig,
			PreRunE: MakePreRunSetKeeperConfig(pc.CmdConfig.Run, fields),
		}

		if pc.CmdConfig.FlagInit != nil {
			if err := pc.CmdConfig.FlagInit(subCmd.Flags()); err != nil {
				s.Logger.Panic(err)
				return err
			}
		}

		for name, desc := range pc.CmdConfig.Fields {
			var secOpt LiteralOrSecretRef
			subCmd.Flags().StringVar(&secOpt.Literal, name, "", desc)
			subCmd.Flags().Var(&flag.Secret{SecretRef: &secOpt.Ref}, name+"-secret", desc+" (set from a secret lookup)")
			fields[name] = &secOpt
		}

		SetCmd.AddCommand(subCmd)
	}
	return nil
}

func RunSetKeeperConfig(cmd *cobra.Command, args []string) {
	if Replacement == nil {
		s.Logger.Panicf("Configuration failed. No replacement configuration was set.")
	}

	keeperName := args[0]
	c := config.Instance()
	kc := c.Keepers[keeperName]
	var wasType string
	if kc != nil {
		was := plugin.Type(kc)
		if was == "" {
			s.Logger.Panicf("Configuration failed. Keeper %q has no type.", keeperName)
		}
	}

	c.Keepers[keeperName] = Replacement

	newType := Replacement["type"].(string)
	if newType == "" {
		s.Logger.Panicf("Configuration failed. Keeper %q has no type.", keeperName)
		return
	}

	if wasType != "" && newType != wasType {
		s.Logger.Panicf("Configuration failed. New keeper type %q does not match old type %q.", newType, wasType)
		return
	}

	if kc != nil && wasType != plugin.Type(kc) {
		s.Logger.Panicf("Configuration failed. New kc type %q does not match old type %q.", newType, wasType)
		return
	}

	err := keeper.CheckConfig(context.Background(), c)
	if err != nil {
		s.Logger.Panicf("Configuration failed. Configuration errors: %v", err)
		return
	}

	err = c.Save(s.ConfigFile)
	if err != nil {
		s.Logger.Panicf("Configuration failed. Error saving configuration: %v", err)
	}
}

func MakePreRunSetKeeperConfig(
	run plugin.CmdFunc,
	fields map[string]*LiteralOrSecretRef,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfgFields := make(map[string]any, len(fields))
		for name, opt := range fields {
			if opt.Literal != "" && opt.Ref.KeeperName != "" {
				return fmt.Errorf("cannot use both --%s and --%s-secret", name, name)
			}

			if opt.Literal != "" {
				cfgFields[name] = opt.Literal
			}

			cfgFields[name] = map[string]any{
				"keeper": opt.Ref.KeeperName,
				"secret": opt.Ref.SecretName,
				"field":  opt.Ref.Field,
			}
		}

		keeperName := args[0]
		repl, err := run(keeperName, cfgFields)
		if err != nil {
			return err
		}

		Replacement = repl

		return nil
	}
}
