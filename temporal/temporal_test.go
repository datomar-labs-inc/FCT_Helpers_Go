package fcttemporal

import (
	"context"
	"fmt"
	"testing"
)

func TestSetupTemporal(t *testing.T) {
	client := SetupTemporal(&TemporalSetupConfig{
		Namespace:            "test_namespace",
		NamespaceDescription: "a test namespace",
		Endpoint:             "localhost:7233",
	})

	searchAttributes, err := client.GetSearchAttributes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(searchAttributes)
}
