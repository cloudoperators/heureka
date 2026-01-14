package common

import (
	"os"

	"github.com/cloudoperators/heureka/internal/util"
)

func GetTestConfig() *util.Config {
	var modelFilePath = "./../../openfga/model/model.fga"

	return &util.Config{
		AuthzOpenFgaApiUrl:    os.Getenv("AUTHZ_FGA_API_URL"),
		AuthzOpenFgaApiToken:  os.Getenv("AUTHZ_FGA_API_TOKEN"),
		AuthzOpenFgaStoreName: os.Getenv("AUTHZ_FGA_STORE_NAME"),
		AuthzModelFilePath:    modelFilePath,
		CurrentUser:           "testuser",
	}
}
