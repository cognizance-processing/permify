package cmd

import (
	"context"
	"os/signal"
	"permify/internal/repositories/postgres"
	PQDatabase "permify/pkg/database/postgres"
	"syscall"

	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/spf13/viper"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"

	"permify/internal"
	"permify/internal/config"
	"permify/internal/engines"
	"permify/internal/factories"
	"permify/internal/keys"
	"permify/internal/repositories"
	"permify/internal/repositories/decorators"
	"permify/internal/servers"
	"permify/internal/services"
	"permify/pkg/cache"
	"permify/pkg/cache/ristretto"
	"permify/pkg/database"
	"permify/pkg/logger"
	"permify/pkg/telemetry"
	"permify/pkg/telemetry/meterexporters"
	"permify/pkg/telemetry/tracerexporters"
)

// NewServeCommand returns a new Cobra command that can be used to run the "permify serve" command.
// The command takes no arguments and runs the serve() function to start the Permify service.
// The command has a short description of what it does.
func NewServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serve the Permify server",
		RunE:  serve(),
		Args:  cobra.NoArgs,
	}
}

// serve is the main function for the "permify serve" command. It starts the Permify service by configuring and starting the necessary components.
// It initializes the configuration, logger, database, tracing and metering components, and creates instances of the necessary engines, services, and decorators.
// It then creates a ServiceContainer and runs it with the given configuration.
// The function uses errgroup to manage the goroutines and gracefully shuts down the service upon receiving a termination signal.
// It returns an error if there is an issue with any of the components or if any goroutine fails.
func serve() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfig()
		if err != nil {
			return err
		}

		if err = viper.Unmarshal(cfg); err != nil {
			return err
		}

		red := color.New(color.FgGreen)
		_, _ = red.Printf(internal.Banner, internal.Version)

		l := logger.New(cfg.Log.Level)

		l.Info("🚀 starting permify service...")

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		if cfg.AutoMigrate {
			err = repositories.Migrate(cfg.Database, l)
			if err != nil {
				l.Fatal(err)
			}
		}

		var db database.Database
		db, err = factories.DatabaseFactory(cfg.Database)
		if err != nil {
			l.Fatal(err)
		}
		defer db.Close()

		// Tracing
		if cfg.Tracer.Enabled {
			var exporter trace.SpanExporter
			exporter, err = tracerexporters.ExporterFactory(cfg.Tracer.Exporter, cfg.Tracer.Endpoint)
			if err != nil {
				l.Fatal(err)
			}

			shutdown := telemetry.NewTracer(exporter)

			defer func() {
				if err = shutdown(context.Background()); err != nil {
					l.Fatal(err)
				}
			}()
		}

		// Garbage collection
		if cfg.DatabaseGarbageCollection.Enable && cfg.Database.Engine != "memory" {
			l.Info("🗑️ starting database garbage collection...")
			gc := postgres.NewGarbageCollector(ctx, db.(*PQDatabase.Postgres), l, cfg.DatabaseGarbageCollection)

			err := gc.Start()
			if err != nil {
				l.Fatal(err)
			}

			defer func() {
				gc.Stop()
			}()
		}

		// Meter
		// meter := telemetry.NewNoopMeter()
		if cfg.Meter.Enabled {
			var exporter metric.Exporter
			exporter, err = meterexporters.ExporterFactory(cfg.Meter.Exporter, cfg.Meter.Endpoint)
			if err != nil {
				l.Fatal(err)
			}

			_, err = telemetry.NewMeter(exporter)
			if err != nil {
				l.Fatal(err)
			}
		}

		// schema cache
		var schemaCache cache.Cache
		schemaCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Schema.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Schema.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		// engines cache keys
		var engineKeyCache cache.Cache
		engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(cfg.Permission.Cache.NumberOfCounters), ristretto.MaxCost(cfg.Permission.Cache.MaxCost))
		if err != nil {
			l.Fatal(err)
		}

		// Repositories
		relationshipReader := factories.RelationshipReaderFactory(db, l)
		relationshipWriter := factories.RelationshipWriterFactory(db, l)
		schemaReader := factories.SchemaReaderFactory(db, l)
		schemaWriter := factories.SchemaWriterFactory(db, l)
		tenantReader := factories.TenantReaderFactory(db, l)
		tenantWriter := factories.TenantWriterFactory(db, l)

		// decorators
		schemaReader = decorators.NewSchemaReaderWithCache(schemaReader, schemaCache)

		// Service
		if cfg.Service.CircuitBreaker {
			relationshipWriter = decorators.NewRelationshipWriterWithCircuitBreaker(relationshipWriter)
			relationshipReader = decorators.NewRelationshipReaderWithCircuitBreaker(relationshipReader)

			schemaWriter = decorators.NewSchemaWriterWithCircuitBreaker(schemaWriter)
			schemaReader = decorators.NewSchemaReaderWithCircuitBreaker(schemaReader)
		}

		// key managers
		checkKeyManager := keys.NewCheckEngineKeys(engineKeyCache)

		// engines
		checkEngine := engines.NewCheckEngine(checkKeyManager, schemaReader, relationshipReader, engines.CheckConcurrencyLimit(cfg.Permission.ConcurrencyLimit))
		linkedEntityEngine := engines.NewLinkedEntityEngine(schemaReader, relationshipReader)
		lookupEntityEngine := engines.NewLookupEntityEngine(checkEngine, linkedEntityEngine, engines.LookupEntityConcurrencyLimit(cfg.Permission.BulkLimit))
		expandEngine := engines.NewExpandEngine(schemaReader, relationshipReader)
		schemaLookupEngine := engines.NewLookupSchemaEngine(schemaReader)

		// Services
		relationshipService := services.NewRelationshipService(relationshipReader, relationshipWriter, schemaReader)
		permissionService := services.NewPermissionService(checkEngine, expandEngine, schemaLookupEngine, lookupEntityEngine)
		schemaService := services.NewSchemaService(schemaWriter, schemaReader)
		tenancyService := services.NewTenancyService(tenantWriter, tenantReader)

		container := servers.ServiceContainer{
			RelationshipService: relationshipService,
			PermissionService:   permissionService,
			SchemaService:       schemaService,
			TenancyService:      tenancyService,
		}

		var g *errgroup.Group
		g, ctx = errgroup.WithContext(ctx)

		g.Go(func() error {
			return container.Run(ctx, &cfg.Server, &cfg.Authn, &cfg.Profiler, l)
		})

		if err = g.Wait(); err != nil {
			l.Error(err)
		}

		return nil
	}
}
