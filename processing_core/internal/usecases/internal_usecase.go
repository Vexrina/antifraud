package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"processing_core/internal/app/model"
	desc "processing_core/pkg/core"
)

type (
	InternalClientRepo interface {
		GetCurrentBalanceTx(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) (int64, error)
		UpdateBalance(ctx context.Context, tx pgx.Tx, clientID uuid.UUID, newBalance int64) error

		AddOperationToHistory(ctx context.Context, tx pgx.Tx, clientID uuid.UUID, operation model.Transaction) error
	}

	AntifraudInternalCheck interface {
		InternalCheck(ctx context.Context, operation model.Transaction) error
	}

	internalUsecase struct {
		commonRepo CommonRepo
		clientRepo InternalClientRepo

		antifraud AntifraudInternalCheck
	}
)

func (u *internalUsecase) Process(ctx context.Context, domainRequest *model.InternalDomainRequest) (*desc.InternalResponse, error) {
	afCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := u.antifraud.InternalCheck(afCtx, *domainRequest.Transaction)
	if err != nil {
		// decline / reject / mark suspicious
		return nil, err
	}


	return nil, nil
}
