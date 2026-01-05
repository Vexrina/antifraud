package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	dbModel "processing_core/generated/proc_core_db/public/model"
	"processing_core/internal/app/model"
	"processing_core/internal/repository"
	desc "processing_core/pkg/core"
)

//go:generate mockgen -source=internal_usecase.go -destination=./mocks/internal_usecase_mock.go -package=mocks
type (
	InternalClientRepo interface {
		GetCurrentBalanceTx(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) (int64, error)
		UpdateBalance(ctx context.Context, tx pgx.Tx, clientID uuid.UUID, newBalance int64) error

		UpsertTransaction(ctx context.Context, tx pgx.Tx, transaction dbModel.TransactionsHistory) error
	}

	AntifraudInternalCheck interface {
		InternalCheck(ctx context.Context, operation *model.InternalDomainRequest) error
	}

	internalUsecase struct {
		commonRepo CommonRepo
		clientRepo InternalClientRepo

		antifraud AntifraudInternalCheck
	}
)

func NewInternalUsecase(
	commonRepo CommonRepo,
	clientRepo InternalClientRepo,
	af AntifraudInternalCheck,
) *internalUsecase {
	return &internalUsecase{
		commonRepo: commonRepo,
		clientRepo: clientRepo,
		antifraud:  af,
	}
}

func (u *internalUsecase) Process(ctx context.Context, domainRequest *model.InternalDomainRequest) (*desc.InternalResponse, error) {
	afCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := u.antifraud.InternalCheck(afCtx, domainRequest)
	// пришла ошибка из антифрода и это НЕ отмена контекста
	if err != nil && !(errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
		return &desc.InternalResponse{
			NewStatus: desc.OperationStatus_Declined,
		}, err
	}

	err = u.commonRepo.Transactional(ctx, func(tx pgx.Tx) error {
		sender := domainRequest.Transaction.SenderID
		receiver := domainRequest.ReceiverId
		amount := domainRequest.Transaction.Amount

		txErr := u.commonRepo.LockClient(ctx, tx, sender)
		if txErr != nil {
			return txErr
		}

		txErr = u.commonRepo.LockClient(ctx, tx, receiver)
		if txErr != nil {
			return txErr
		}

		senderBalance, txErr := u.clientRepo.GetCurrentBalanceTx(ctx, tx, sender)
		if txErr != nil {
			return txErr
		}
		if senderBalance < amount {
			return fmt.Errorf("LIMIT OVERFLOW")
		}
		txErr = u.clientRepo.UpdateBalance(ctx, tx, sender, senderBalance-amount)
		if txErr != nil {
			return txErr
		}

		receiverBalance, txErr := u.clientRepo.GetCurrentBalanceTx(ctx, tx, receiver)
		if txErr != nil {
			return txErr
		}
		txErr = u.clientRepo.UpdateBalance(ctx, tx, receiver, receiverBalance+amount)
		if txErr != nil {
			return txErr
		}

		txErr = u.clientRepo.UpsertTransaction(ctx, tx, repository.MapInternalDomainToTransaction(*domainRequest))
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return &desc.InternalResponse{
			NewStatus: desc.OperationStatus_Declined,
		}, err
	}

	return &desc.InternalResponse{
		NewStatus: desc.OperationStatus_Approved,
	}, nil
}
