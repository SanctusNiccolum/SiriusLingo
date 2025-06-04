package db

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	insertTag     = "insert"
	selectTag     = "db"
	updateTag     = "update"
	updateAuthTag = "updateAuth"
)

func colNamesWithPref(cols []string, pref string) []string {
	prefCols := make([]string, len(cols))
	copy(prefCols, cols)
	sort.Strings(prefCols)
	if pref == "" {
		return prefCols
	}

	for i := range prefCols {
		if !strings.Contains(prefCols[i], ".") {
			prefCols[i] = fmt.Sprintf("%s.%s", pref, prefCols[i])
		}
	}
	return prefCols
}

func acquireHealthyConn(ctx context.Context, logger *zap.Logger, runner *pgxpool.Pool) (*pgxpool.Conn, error) {
	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		conn, err := runner.Acquire(ctx)
		if err != nil {
			logger.Warn("Failed to acquire connection",
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", maxAttempts),
				zap.Error(err),
			)
			if attempt == maxAttempts-1 {
				return nil, fmt.Errorf("failed to acquire connection after %d attempts: %w", maxAttempts, err)
			}
			continue
		}

		err = conn.Ping(ctx)
		if err != nil {
			logger.Warn("Connection is not healthy",
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", maxAttempts),
				zap.Error(err),
			)
			conn.Release()
			if attempt == maxAttempts-1 {
				return nil, fmt.Errorf("failed to find healthy connection after %d attempts: %w", maxAttempts, err)
			}
			continue
		}

		logger.Debug("Acquired healthy connection", zap.Int("attempt", attempt+1))
		return conn, nil
	}
	return nil, fmt.Errorf("failed to find healthy connection after %d attempts", maxAttempts)
}
