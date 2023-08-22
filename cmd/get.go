package cmd

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/set"
	"github.com/zostay/go-std/slices"
	"gopkg.in/yaml.v3"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/secrets"
)

var (
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a secret",
		Run:   RunGet,
	}

	id           string
	name         string
	flds         []string
	showPassword bool
	output       string
	one          bool
	envPrefix    string
)

func init() {
	getCmd.Flags().StringVar(&keeperName, "keeper", "", "The name of the secret keeper to use")
	getCmd.Flags().StringVar(&id, "id", "", "The ID of the secret to get")
	getCmd.Flags().StringVar(&name, "name", "", "The name of the secret to get")
	getCmd.Flags().StringSliceVar(&flds, "fields", []string{}, "The fields to display")
	getCmd.Flags().BoolVar(&showPassword, "show-password", false, "Show the password in the output")
	getCmd.Flags().StringVarP(&output, "output", "o", "pretty", "Output format (pretty, yaml, json, env, password)")
	getCmd.Flags().BoolVarP(&one, "one", "1", false, "If multiple secrets found, print only the first found")
	getCmd.Flags().StringVar(&envPrefix, "env-prefix", "", "The prefix to use when output is env")
}

func RunGet(cmd *cobra.Command, args []string) {
	if name != "" && id != "" {
		s.Logger.Panic("Cannot specify both --id and --name.")
	}

	if name == "" && id == "" {
		s.Logger.Panic("Must specify either --id or --name.")
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

	if one && len(secs) > 0 {
		secs = secs[0:1]
	}

	switch output {
	case "env":
		if one {
			ms := convertToMap(secs)
			for k, v := range ms[0] {
				s.Logger.Printf("export %s%s=%s", strings.ToUpper(envPrefix), strings.ToUpper(k), v)
			}
		} else {
			s.Logger.Panic("Cannot enable --output=env without --one")
		}
	case "json":
		ms := convertToMap(secs)
		sb := &strings.Builder{}
		enc := json.NewEncoder(sb)
		enc.SetIndent("", "  ")
		if one {
			if err := enc.Encode(ms[0]); err != nil {
				s.Logger.Panic(err)
			}
		} else if err := enc.Encode(map[string]any{"secrets": ms}); err != nil {
			s.Logger.Panic(err)
		}
		s.Logger.Print(sb.String())
	case "password":
		if showPassword {
			for _, sec := range secs {
				s.Logger.Print(sec.Password())
			}
		} else {
			s.Logger.Panic("Cannot enable --output=password without --show-password")
		}
	case "pretty":
		for _, sec := range secs {
			s.PrintSecret(sec, showPassword, flds...)
		}
	case "yaml":
		ms := convertToMap(secs)
		sb := &strings.Builder{}
		enc := yaml.NewEncoder(sb)
		enc.SetIndent(2)
		if one {
			if err := enc.Encode(ms[0]); err != nil {
				s.Logger.Panic(err)
			}
		} else if err := enc.Encode(map[string]any{"secrets": ms}); err != nil {
			s.Logger.Panic(err)
		}
		s.Logger.Print(sb.String())
	}
}

func convertToMap(secs []secrets.Secret) []map[string]any {
	fldSet := set.New[string](slices.Map(flds, strings.ToLower)...)

	ms := make([]map[string]any, len(secs))
	for i, sec := range secs {
		ms[i] = make(map[string]any, 10)
		if fldSet.Len() == 0 || fldSet.Contains("id") {
			ms[i]["id"] = sec.ID()
		}

		if fldSet.Len() == 0 || fldSet.Contains("name") {
			ms[i]["name"] = sec.Name()
		}

		if fldSet.Len() == 0 || fldSet.Contains("username") {
			ms[i]["username"] = sec.Username()
		}

		if showPassword && (fldSet.Len() == 0 || fldSet.Contains("password")) {
			ms[i]["password"] = sec.Password()
		}

		if fldSet.Len() == 0 || fldSet.Contains("location") {
			ms[i]["location"] = sec.Location()
		}

		if fldSet.Len() == 0 || fldSet.Contains("url") {
			url := ""
			if sec.Url() != nil {
				url = sec.Url().String()
			}
			ms[i]["url"] = url
		}

		if fldSet.Len() == 0 || fldSet.Contains("last-modified") {
			ms[i]["last-modified"] = sec.LastModified()
		}

		if fldSet.Len() == 0 || fldSet.Contains("type") {
			ms[i]["type"] = sec.Type()
		}

		var fields map[string]string
		for k, v := range sec.Fields() {
			if fldSet.Len() == 0 || fldSet.Contains(strings.ToLower(k)) {
				if fields == nil {
					fields = make(map[string]string, len(sec.Fields()))
				}
				fields[k] = v
			}
		}

		if fields != nil {
			ms[i]["fields"] = fields
		}
	}

	return ms
}
