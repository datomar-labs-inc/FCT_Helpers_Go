package fcttemporal

import (
	"context"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"go.uber.org/zap"
	"testing"
)

func TestSetupTemporal(t *testing.T) {
	logger := lggr.NewTest(t)

	client := SetupTemporal(&TemporalSetupConfig{
		Namespace:            "test_namespace",
		NamespaceDescription: "a test namespace",
		Endpoint:             "localhost:7233",
	}, logger)

	searchAttributes, err := client.GetSearchAttributes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	logger.Info("search attributes", zap.Any("search_attributes", searchAttributes))
}
