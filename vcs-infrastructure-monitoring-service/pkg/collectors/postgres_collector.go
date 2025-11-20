package collectors

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type IPostgreSQLCollector interface {
	CollectMetrics(ctx context.Context, host string, port int, username, password, database string) (*PostgreSQLMetrics, error)
}

type PostgreSQLMetrics struct {
	ActiveConnections int64
	TotalConnections  int64
	Transactions      int64
	Commits           int64
	Rollbacks         int64
	BlocksRead        int64
	BlocksHit         int64
	TuplesReturned    int64
	TuplesFetched     int64
	TuplesInserted    int64
	TuplesUpdated     int64
	TuplesDeleted     int64
	ReplicationLag    int64
}

type postgresCollector struct {
	logger logger.ILogger
}

func NewPostgreSQLCollector(logger logger.ILogger) IPostgreSQLCollector {
	return &postgresCollector{
		logger: logger,
	}
}

func (pc *postgresCollector) CollectMetrics(ctx context.Context, host string, port int, username, password, database string) (*PostgreSQLMetrics, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, username, password, database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		pc.logger.Error("failed to connect to postgres", zap.Error(err))
		return nil, err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		pc.logger.Error("failed to ping postgres", zap.Error(err))
		return nil, err
	}

	metrics := &PostgreSQLMetrics{}

	row := db.QueryRowContext(ctx, `
		SELECT 
			(SELECT count(*) FROM pg_stat_activity WHERE state = 'active'),
			(SELECT count(*) FROM pg_stat_activity),
			(SELECT sum(xact_commit + xact_rollback) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(xact_commit) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(xact_rollback) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(blks_read) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(blks_hit) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(tup_returned) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(tup_fetched) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(tup_inserted) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(tup_updated) FROM pg_stat_database WHERE datname = $1),
			(SELECT sum(tup_deleted) FROM pg_stat_database WHERE datname = $1)
	`, database)

	err = row.Scan(
		&metrics.ActiveConnections,
		&metrics.TotalConnections,
		&metrics.Transactions,
		&metrics.Commits,
		&metrics.Rollbacks,
		&metrics.BlocksRead,
		&metrics.BlocksHit,
		&metrics.TuplesReturned,
		&metrics.TuplesFetched,
		&metrics.TuplesInserted,
		&metrics.TuplesUpdated,
		&metrics.TuplesDeleted,
	)
	if err != nil {
		pc.logger.Error("failed to collect postgres metrics", zap.Error(err))
		return nil, err
	}

	var isReplica bool
	err = db.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&isReplica)
	if err != nil {
		pc.logger.Warn("failed to check replication status", zap.Error(err))
	}

	if isReplica {
		var lagBytes int64
		err = db.QueryRowContext(ctx, "SELECT pg_wal_lsn_diff(pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn())").Scan(&lagBytes)
		if err != nil {
			pc.logger.Warn("failed to get replication lag", zap.Error(err))
		} else {
			metrics.ReplicationLag = lagBytes
		}
	}

	return metrics, nil
}
