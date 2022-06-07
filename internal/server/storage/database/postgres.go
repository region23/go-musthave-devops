package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
)

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

// при инициализации базы данных проверить, есть ли таблица metrics
// если нет, то создать
