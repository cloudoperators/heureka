// mock_oidc_provider/main.go
package main

import (
	"log"
	"os"

	"github.com/cloudoperators/heureka/pkg/util"
)

func main() {
	url := os.Getenv("OIDC_PROVIDER_URL")
	if url == "" {
		log.Fatal("Could not start OIDC provider. OIDC_PROVIDER_URL is not set.")
	}
    oidcProvider := util.NewOidcProvider(url)
    oidcProvider.StartForeground()
}
