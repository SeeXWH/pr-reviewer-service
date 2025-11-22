//go:build integration

package tests

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgresContainer() (*postgres.PostgresContainer, func(), error) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithStartupTimeout(5*time.Minute),
		),
	)

	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		pgContainer.Terminate(ctx)
	}

	return pgContainer, cleanup, nil
}
