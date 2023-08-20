package main

import (
	"github.com/zostay/ghost/cmd"
	_ "github.com/zostay/ghost/pkg/secrets/http"
	_ "github.com/zostay/ghost/pkg/secrets/human"
	_ "github.com/zostay/ghost/pkg/secrets/keepass"
	_ "github.com/zostay/ghost/pkg/secrets/keyring"
	_ "github.com/zostay/ghost/pkg/secrets/lastpass"
	_ "github.com/zostay/ghost/pkg/secrets/low"
	_ "github.com/zostay/ghost/pkg/secrets/memory"
	_ "github.com/zostay/ghost/pkg/secrets/policy"
	_ "github.com/zostay/ghost/pkg/secrets/router"
	_ "github.com/zostay/ghost/pkg/secrets/seq"
)

func main() {
	cmd.Execute()
}
