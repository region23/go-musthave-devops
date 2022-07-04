package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/region23/go-musthave-devops/internal/serializers"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

type InDatabase struct {
	mu     sync.Mutex
	dbpool *pgxpool.Pool
	key    string
}

func NewInDatabase(dbpool *pgxpool.Pool, key string) storage.Repository {
	return &InDatabase{
		dbpool: dbpool,
		key:    key,
	}
}

// проверяем есть ли соединение с базой данных
func Ping(dbpool *pgxpool.Pool) error {
	if dbpool == nil {
		return errors.New("connection is nil")
	}

	err := dbpool.Ping(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// При инициализации базы данных проверить, есть ли таблица metrics.
// Если её нет, то создать.
func InitDB(dbpool *pgxpool.Pool) error {
	query := `CREATE TABLE IF NOT EXISTS metrics (
		id VARCHAR(50) UNIQUE,
		metric_type VARCHAR(10) not null,
		delta BIGINT DEFAULT NULL,
		gauge double precision DEFAULT NULL,
		hash VARCHAR(64) DEFAULT NULL
	  );`

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	_, err := dbpool.Exec(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Error when creating product table")
		return err
	}

	//rows := res.RowsAffected()
	//log.Printf("Rows affected when creating table: %d", rows)
	return nil
}

// извлекает метрику из базы данных
func (storage *InDatabase) Get(key string) (*serializers.Metrics, error) {
	row := storage.dbpool.QueryRow(context.Background(),
		`SELECT id, metric_type, delta, gauge, hash FROM metrics WHERE id = $1`,
		key)

	var metric serializers.Metrics

	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value, &metric.Hash)

	switch err {
	case nil:
		return &metric, nil
	case pgx.ErrNoRows:
		return nil, pgx.ErrNoRows
	default:
		return nil, err
	}

}

func (storage *InDatabase) Put(metric *serializers.Metrics) error {
	// если это counter, то извлекаем из базы последнее значение счетчика и увеличиваем его на значение метрики
	if metric.MType == "counter" {
		storage.mu.Lock()
		defer storage.mu.Unlock()
		metricFromDB, err := storage.Get(metric.ID)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}

		if err == nil {
			*metric.Delta = *metricFromDB.Delta + *metric.Delta

			// обновим хэш метрики
			if storage.key != "" {
				metric.Hash = serializers.Hash(metric.MType, metric.ID, fmt.Sprintf("%d", *metric.Delta), storage.key)
			}
		}
	}

	_, err := storage.dbpool.Exec(context.Background(),
		`INSERT INTO metrics (id, metric_type, delta, gauge, hash) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id)
	DO UPDATE 
	SET delta = $3, gauge = $4, hash = $5;`,
		metric.ID,
		metric.MType,
		metric.Delta,
		metric.Value,
		metric.Hash)

	if err != nil {
		log.Error().Err(err).Msg("Unable to INSERT metric to DB")
		return err
	}

	return nil
}

func (storage *InDatabase) All() (map[string]serializers.Metrics, error) {
	rows, err := storage.dbpool.Query(context.Background(),
		`SELECT id, metric_type, delta, gauge, hash FROM metrics`)

	if err != nil {
		return nil, err
	}

	metrics := make(map[string]serializers.Metrics)

	for rows.Next() {
		var metric serializers.Metrics
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value, &metric.Hash)
		if err != nil {
			return nil, err
		}

		metrics[metric.ID] = metric
	}

	return metrics, rows.Err()

}

func (storage *InDatabase) UpdateAll(m map[string]serializers.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := storage.dbpool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = storage.deleteAll(tx)
	if err != nil {
		return err
	}

	rows := [][]interface{}{}

	for _, metric := range m {
		rows = append(rows, []interface{}{metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash})
	}

	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"metrics"},
		[]string{"id", "metric_type", "delta", "gauge", "hash"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return err
	}

	_, err = storage.dbpool.CopyFrom(ctx,
		pgx.Identifier{"metrics"},
		[]string{"id", "metric_type", "delta", "gauge", "hash"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)

}

// Удаляет все записи из таблицы metrics
func (storage *InDatabase) deleteAll(tx pgx.Tx) error {
	ct, err := tx.Exec(context.Background(), `DELETE FROM metrics`)

	if err != nil {
		log.Error().Err(err).Msg("Unable to DELETE metrics from DB")
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("no rows to delete")
	}

	return nil
}
