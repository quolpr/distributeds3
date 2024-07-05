package pg

import (
	"github.com/jackc/pgx/v5"
)

// QuerierTX расширение интерфейса pg.Querier, поддерживающее транзакции.
type QuerierTX interface {
	Querier
	WithTx(tx pgx.Tx) Querier
}

type TxQueries struct {
	*Queries
}

func NewTxQueries(base *Queries) *TxQueries {
	return &TxQueries{base}
}

//nolint:ireturn
func (_d TxQueries) WithTx(tx pgx.Tx) Querier {
	return NewTxQueries(_d.Queries.WithTx(tx))
}
