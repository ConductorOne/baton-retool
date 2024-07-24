package main

import (
	"context"
	"fmt"
	"os"

	configschema "github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/conductorone/baton-retool/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := configschema.DefineConfiguration(ctx, "baton-retool", getConnector, configuration)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	connString := v.GetString(ConnectionString.FieldName)
	skipPages := v.GetBool(SkipPages.FieldName)
	skipResources := v.GetBool(SkipResources.FieldName)
	skipDisabledUsers := v.GetBool(SkipDisabledUsers.FieldName)

	cb, err := connector.New(ctx, connString, skipPages, skipResources, skipDisabledUsers)
	if err != nil {
		l.Error("error creating connector builder", zap.Error(err))
		return nil, err
	}

	c, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector from connector builder", zap.Error(err))
		return nil, err
	}

	return c, nil
}
