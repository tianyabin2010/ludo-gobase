package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"time"
)

func NewMysqlClientPool(addr string, maxConn, maxIdle int) (*sql.DB, error) {
	ret, err := sql.Open("mysql", addr)
	if err != nil {
		log.Error().Err(err).
			Msgf("NewMysqlClientPool sql.Open error")
		return nil, err
	}
	ret.SetConnMaxLifetime(time.Minute * 3)
	ret.SetMaxOpenConns(maxConn)
	ret.SetMaxIdleConns(maxIdle)
	return ret, nil
}