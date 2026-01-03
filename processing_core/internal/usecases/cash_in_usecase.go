package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx"

	"processing_core/internal/app/model"
	desc "processing_core/pkg/core"
)

type (
	commonRepo interface {
		Transactional(ctx context.Context, f func(tx pgx.Tx) error) error
		LockClient(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) error
	}

	clientRepo interface {
		GetCurrentBalance(ctx context.Context, clientID uuid.UUID) (int64, error)
		UpdateBalance(ctx context.Context, clientID uuid.UUID, newBalance int64) error
	}

	cashInUsecase struct {
		commonRepo commonRepo
		clientRepo clientRepo
	}
)

func NewCashInUsecase(commonRepo commonRepo, clientRepo clientRepo) *cashInUsecase {
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
		clientBalance, txErr := u.clientRepo.GetCurrentBalance(ctx, clientID)
		if txErr != nil {
			return txErr
		}
		return u.clientRepo.UpdateBalance(ctx, clientID, clientBalance+domainRequest.Transaction.Amount)
	})
	if err != nil{
		return nil, err
	}

	return &desc.CashInResponse{
		NewStatus: desc.OperationStatus_Approved,
	}, err
}
