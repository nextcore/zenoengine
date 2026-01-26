package worker

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"zeno/pkg/dbmanager"
)

type DBQueue struct {
	dbMgr    *dbmanager.DBManager
	connName string
}

func NewDBQueue(dbMgr *dbmanager.DBManager, connName string) *DBQueue {
	return &DBQueue{
		dbMgr:    dbMgr,
		connName: connName,
	}
}

func (q *DBQueue) Push(ctx context.Context, queue string, payload []byte) error {
	db := q.dbMgr.GetConnection(q.connName)
	dialect := q.dbMgr.GetDialect(q.connName)

	// Gunakan parameterized query standard
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (%s, %s, %s, %s)",
		dialect.QuoteIdentifier("jobs"),
		dialect.QuoteIdentifier("queue"),
		dialect.QuoteIdentifier("payload"),
		dialect.QuoteIdentifier("status"),
		dialect.QuoteIdentifier("created_at"),
		dialect.Placeholder(1),
		dialect.Placeholder(2),
		dialect.Placeholder(3),
		dialect.Placeholder(4))

	// SQLite support datetime standard time.Time go
	_, err := db.ExecContext(ctx, query, queue, payload, "pending", time.Now())
	return err
}

func (q *DBQueue) Pop(ctx context.Context, queues []string) (string, []byte, error) {
	db := q.dbMgr.GetConnection(q.connName)
	dialect := q.dbMgr.GetDialect(q.connName)

	// Polling Interval
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-ticker.C:
			// OPTIMISTIC LOCKING STRATEGY
			// 1. Ambil 1 job pending
			var id int
			var queueName string
			var payload []byte

			// Build Filter Queue
			whereQueue := ""
			args := []interface{}{}
			if len(queues) > 0 {
				placeholders := make([]string, len(queues))
				for i, qName := range queues {
					placeholders[i] = dialect.Placeholder(i + 1)
					args = append(args, qName)
				}
				whereQueue = fmt.Sprintf("AND %s IN (%s)", dialect.QuoteIdentifier("queue"), strings.Join(placeholders, ","))
			}

			querySelect := fmt.Sprintf("SELECT %s, %s, %s FROM %s WHERE %s = %s %s ORDER BY %s ASC %s",
				dialect.QuoteIdentifier("id"),
				dialect.QuoteIdentifier("queue"),
				dialect.QuoteIdentifier("payload"),
				dialect.QuoteIdentifier("jobs"),
				dialect.QuoteIdentifier("status"),
				dialect.Placeholder(len(args)+1),
				whereQueue,
				dialect.QuoteIdentifier("created_at"),
				dialect.Limit(1, 0))

			fullArgs := append(args, "pending")
			err := db.QueryRowContext(ctx, querySelect, fullArgs...).Scan(&id, &queueName, &payload)
			if err == sql.ErrNoRows {
				continue // Tidak ada tugas, tunggu tick berikutnya
			} else if err != nil {
				return "", nil, err // Error DB serius
			}

			// 2. Coba claim (Atomic Update)
			updateQuery := fmt.Sprintf("UPDATE %s SET %s = %s, %s = %s WHERE %s = %s AND %s = %s",
				dialect.QuoteIdentifier("jobs"),
				dialect.QuoteIdentifier("status"), dialect.Placeholder(1),
				dialect.QuoteIdentifier("processed_at"), dialect.Placeholder(2),
				dialect.QuoteIdentifier("id"), dialect.Placeholder(3),
				dialect.QuoteIdentifier("status"), dialect.Placeholder(4))

			res, err := db.ExecContext(ctx, updateQuery, "processing", time.Now(), id, "pending")
			if err != nil {
				return "", nil, err
			}

			affected, _ := res.RowsAffected()
			if affected == 1 {
				return queueName, payload, nil
			}

			continue
		}
	}
}

func (q *DBQueue) Close() error {
	return nil // DB connections closed by main app
}
