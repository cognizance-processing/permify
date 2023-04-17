package postgres

import (
	"context"
	"database/sql"
	"permify/internal/config"
	"permify/internal/repositories"
	"permify/internal/repositories/postgres/utils"
	db "permify/pkg/database/postgres"
	"permify/pkg/logger"
	base "permify/pkg/pb/base/v1"
	"time"

	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// GarbageCollector - Structure for GarbageCollector
type GarbageCollector struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
	// logger
	logger logger.Interface
	// context to manage goroutines and cancellation
	ctx context.Context
	// errgroup for managing multiple goroutines
	g *errgroup.Group
	// limit for concurrent permission checks
	concurrencyLimit int
	// interval for garbage collection
	interval time.Duration
	// timeout for garbage collection
	timeout time.Duration
	// window for garbage collection
	window time.Duration
}

// NewGarbageCollector creates a new GarbageCollector instance.
// ctx: context for managing goroutines and cancellation
// concurrencyLimit: the maximum number of concurrent garbage collection
func NewGarbageCollector(ctx context.Context, db *db.Postgres, logger logger.Interface, cfg config.DatabaseGarbageCollection) *GarbageCollector {
	return &GarbageCollector{
		g:                &errgroup.Group{},
		concurrencyLimit: cfg.NumberOfThreads,
		interval:         cfg.Interval,
		timeout:          cfg.Timeout,
		window:           cfg.Window,
		txOptions:        sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false},
		database:         db,
		logger:           logger,
		ctx:              ctx,
	}
}

// Start begins processing permission check requests from the RequestChan.
// It starts an errgroup that manages multiple goroutines for garbage collector check.
func (c *GarbageCollector) Start() error {
	c.g.Go(func() error {
		sem := semaphore.NewWeighted(int64(c.concurrencyLimit))
		// for loop time ticker

		// tracer start
		ctx, span := tracer.Start(c.ctx, "garbage-collector.start")
		defer span.End()

		ticker := time.NewTicker(c.interval)
		for _ = range ticker.C {
			select {
			case <-ctx.Done():
				c.logger.Info("garbage collector stopped due to timeout")
				return nil
			default:
				c.logger.Info("garbage collector started")
				// acquire a semaphore before processing a request
				if err := sem.Acquire(c.ctx, 1); err != nil {
					return err
				}

				tenants, err := c.getTenants(ctx)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return err
				}

				// run the permission check in a separate goroutine
				c.g.Go(func() error {
					defer sem.Release(1)

					for i := range tenants {
						err := c.executeCollector(ctx, tenants[i].Id)
						if err != nil {
							span.RecordError(err)
							span.SetStatus(codes.Error, err.Error())
							c.logger.Error("garbage collector failed for tenant: " + tenants[i].Id + " with error: " + err.Error())

							return err
						}
						c.logger.Info("garbage collector finished for tenant: " + tenants[i].Id)
					}

					time.Sleep(c.timeout)

					return nil
				})
			}
		}
		// wait for all remaining semaphore resources to be released
		if err := sem.Acquire(c.ctx, int64(c.concurrencyLimit)); err != nil {
			return err

		}
		return nil
	})

	return nil
}

// Stop stops input by closing the GarbageCollector.
func (c *GarbageCollector) Stop() {
	c.ctx.Done()
}

// Wait waits for all goroutines in the errgroup to finish.
// Returns an error if any of the goroutines encounter an error.
func (c *GarbageCollector) Wait() error {
	if err := c.g.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *GarbageCollector) getTenants(ctx context.Context) ([]*base.Tenant, error) {
	// get all tenants
	builder := c.database.Builder.Select("id, name, created_at").From(TenantsTable)

	query, args, err := builder.ToSql()
	if err != nil {

		return nil, err
	}

	var rows *sql.Rows
	rows, err = c.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// close rows
	defer rows.Close()

	// iterate over rows and convert to tenant
	tenants := make([]*base.Tenant, 0)

	for rows.Next() {
		sd := repositories.Tenant{}
		err = rows.Scan(&sd.ID, &sd.Name, &sd.CreatedAt)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, sd.ToTenant())
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tenants, nil
}

func (c *GarbageCollector) executeCollector(ctx context.Context, tenantID string) error {
	garbageQuery := utils.GarbageCollectQuery(c.window, tenantID)

	garbageSQL, garbageQueryArgs, err := garbageQuery.ToSql()
	if err != nil {

		return err
	}

	_, err = c.database.DB.QueryContext(ctx, garbageSQL, garbageQueryArgs...)
	if err != nil {
		return err
	}

	return nil
}
