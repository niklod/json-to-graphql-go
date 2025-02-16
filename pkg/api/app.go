package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/niklod/json-to-graphql-go/internal/builder"
	"github.com/niklod/json-to-graphql-go/internal/field"
	"github.com/niklod/json-to-graphql-go/pkg/data"
	"github.com/niklod/json-to-graphql-go/pkg/handler"
	"github.com/niklod/json-to-graphql-go/pkg/resolver"
)

type App struct {
	Handler         http.Handler
	internalHandler *handler.GraphQLHandler
	jsonProvider    JsonProvider
	schemaBuilder   SchemaBuilder
	resolver        Resolver
	logger          *slog.Logger
}

type Config struct {
	JSONProvider JsonProvider
	Resolver     Resolver
	SchemBuilder SchemaBuilder
}

func New(config Config) (*App, error) {
	jsonProvider := config.JSONProvider
	if jsonProvider == nil {
		jsonProvider = data.NewJSONProvider()
	}

	rawJson, err := jsonProvider.GetRawJson()
	if err != nil {
		return nil, err
	}

	dataResolver := config.Resolver
	if dataResolver == nil {
		dataResolver = resolver.NewJSONResolver(rawJson)
	}

	fieldFactory, err := field.NewDefaultFieldFactory(field.Config{
		Resolver: dataResolver,
	})
	if err != nil {
		return nil, err
	}

	schemaBuilder := config.SchemBuilder
	if schemaBuilder == nil {
		schemaBuilder = builder.NewGraphQLSchemaBuilder(fieldFactory)
	}

	handler := handler.NewGraphQLHandler(schemaBuilder)

	return &App{
		internalHandler: handler,
		Handler:         handler,
		jsonProvider:    jsonProvider,
		schemaBuilder:   schemaBuilder,
		resolver:        dataResolver,
		logger:          slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}, nil
}

func (a *App) SchemaUpdate() error {
	structuredJson, err := a.jsonProvider.GetJsonData()
	if err != nil {
		return err
	}

	rawJson, err := a.jsonProvider.GetRawJson()
	if err != nil {
		return err
	}

	schema, err := a.schemaBuilder.BuildSchema(structuredJson)
	if err != nil {
		return err
	}

	a.internalHandler.UpdateSchema(schema)
	a.resolver.UpdateJsonData(rawJson)

	a.logger.Debug("schema updated")

	return nil
}

func (a *App) StartBackgroundSchemaUpdate(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				a.logger.Debug("context done, stopping schema update")

				return
			case <-ticker.C:
				if err := a.SchemaUpdate(); err != nil {
					a.logger.Error("failed to update schema", slog.Any("error", err))
				}
			}
		}
	}()
}
