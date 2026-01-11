package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	dbModel "processing_core/generated/proc_core_db/public/model"
	"processing_core/internal/app/model"
	"processing_core/internal/repository"
	desc "processing_core/pkg/core"
)

//go:generate mockgen -source=cash_in_usecase.go -destination=./mocks/cash_in_usecase_mock.go -package=mocks
type (
	CommonRepo interface {
		Transactional(ctx context.Context, f func(tx pgx.Tx) error) error
		LockClient(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) error
	}

	CashInClientRepo interface {
		GetCurrentBalanceTx(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) (int64, error)
		UpdateBalance(ctx context.Context, tx pgx.Tx, clientID uuid.UUID, newBalance int64) error
		UpsertTransaction(ctx context.Context, tx pgx.Tx, transaction dbModel.TransactionsHistory) error
	}

	cashInUsecase struct {
		commonRepo CommonRepo
		clientRepo CashInClientRepo
	}
)

func NewCashInUsecase(commonRepo CommonRepo, clientRepo CashInClientRepo) *cashInUsecase {
	return &cashInUsecase{
		commonRepo: commonRepo,
		clientRepo: clientRepo,
	}
}

func (u *cashInUsecase) Process(ctx context.Context, domainRequest *model.CashInDomainRequest) (*desc.CashInResponse, error) {
	err := u.commonRepo.Transactional(ctx, func(tx pgx.Tx) error {
		clientID := domainRequest.Transaction.SenderID
		txErr := u.commonRepo.LockClient(ctx, tx, clientID)
		if txErr != nil {
			return txErr
		}

		clientBalance, txErr := u.clientRepo.GetCurrentBalanceTx(ctx, tx, clientID)
		if txErr != nil {
			if !errors.Is(txErr, pgx.ErrNoRows) {
				return txErr
			}
		}

		txErr = u.clientRepo.UpdateBalance(ctx, tx, clientID, clientBalance+domainRequest.Transaction.Amount)
		if txErr != nil {
			return txErr
		}

		txErr = u.clientRepo.UpsertTransaction(
			ctx,
			tx,
			repository.MapCashInDomainToTransaction(*domainRequest),
		)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return &desc.CashInResponse{
			NewStatus: desc.OperationStatus_Declined,
		}, err
	}

	return &desc.CashInResponse{
		NewStatus: desc.OperationStatus_Approved,
	}, err
}
