package repository

import (
	"context"
	"fmt"
	"time"

	logger "github.com/bllooop/votingbot/pkg/logging"
	"github.com/tarantool/go-tarantool/v2"
)

type Config struct {
	Host     string
	Username string
	Password string
	Port     string
}

func NewTarantoolDB(cfg Config) (*tarantool.Connection, error) {
	logger.Log.Info().Msg("Подключение к базе данных")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	logger.Log.Debug().Any("address", address).Msg("Адрес подключения")
	dialer := tarantool.NetDialer{
		//"tarantool:3301"
		Address:  address,
		User:     cfg.Username,
		Password: cfg.Password,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}
	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
