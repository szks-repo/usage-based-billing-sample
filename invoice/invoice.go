package invoice

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/szks-repo/usage-based-billing-sample/pkg/now"
)

type InvoiceMaker struct {
	dbConn *sql.DB
}

func NewInvoiceMaker(
	dbConn *sql.DB,
) *InvoiceMaker {
	return &InvoiceMaker{
		dbConn: dbConn,
	}
}

func (i *InvoiceMaker) CreateInvoiceDaily(ctx context.Context) {
	accountIds, err := i.listAccountIds(ctx, now.FromContext(ctx))
	if err != nil {
		slog.Error("Failed to listAccountIds", "error", err)
	}

	slog.Info("target accountIds", "accountIds", accountIds)
	if len(accountIds) == 0 {
		return
	}

	// Recocile Usage -> Apply Free Credit And Craete Invoice -> Notify
}

func (i *InvoiceMaker) listAccountIds(ctx context.Context, t time.Time) ([]int64, error) {
	cutoff := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	rows, err := i.dbConn.QueryContext(ctx, `SELECT a.id FROM account a JOIN account_contract ac ON a.id = ac.account_id WHERE ac.estimated_to = ?`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accountIds []int64
	for rows.Next() {
		var accountId int64
		if err := rows.Scan(&accountId); err != nil {
			return nil, err
		}
		accountIds = append(accountIds, accountId)
	}

	return accountIds, nil
}
