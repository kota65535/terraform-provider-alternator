package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerFactories = map[string]func() (*schema.Provider, error){
	"alternator": func() (*schema.Provider, error) {
		os.Setenv("ALTERNATOR_HOST", "localhost:23306")
		os.Setenv("ALTERNATOR_DIALECT", "mysql")
		os.Setenv("ALTERNATOR_USER", "root")
		return New(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
