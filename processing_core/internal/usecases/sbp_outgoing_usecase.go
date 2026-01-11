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

//go:generate mockgen -source=sbp_outgoing_usecase.go -destination=./mocks/sbp_outgoing_usecase_mock.go -package=mocks
type (
	SbpOutgoingClientRepo interface {
		GetCurrentBalanceTx(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) (int64, error)
		UpdateBalance(ctx context.Context, tx pgx.Tx, clientID uuid.UUID, newBalance int64) error

		UpsertTransaction(ctx context.Context, tx pgx.Tx, transaction dbModel.TransactionsHistory) error
	}

	AntifraudSbpOutgoingCheck interface {
		SbpOutgoingCheck(ctx context.Context, operation *model.SbpOutgoingDomainRequest) error
	}

	SbpIntegrationInterface interface {
		// считаем что это асинхронщина, которая при падении СБП не даст списать деньги и если что их вернет на место физически.
		ToAnotherBank(ctx context.Context, operation model.Transaction)
	}

	sbpOutgoingUsecase struct {
		commonRepo CommonRepo
		clientRepo SbpOutgoingClientRepo

		antifraud      AntifraudSbpOutgoingCheck
		sbpIntegration SbpIntegrationInterface
	}
)

func NewSbpOutgoingUsecase(
	commonRepo CommonRepo,
	clientRepo SbpOutgoingClientRepo,
	af AntifraudSbpOutgoingCheck,
	sbpIntegration SbpIntegrationInterface,
) *sbpOutgoingUsecase {
	return &sbpOutgoingUsecase{
		commonRepo:     commonRepo,
		clientRepo:     clientRepo,
		antifraud:      af,
		sbpIntegration: sbpIntegration,
	}
}

func (u *sbpOutgoingUsecase) Process(ctx context.Context, domainRequest *model.SbpOutgoingDomainRequest) (*desc.SbpOutgoingResponse, error) {
	afCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := u.antifraud.SbpOutgoingCheck(afCtx, domainRequest)
	// пришла ошибка из антифрода и это НЕ отмена контекста
	if err != nil && !(errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
		return &desc.SbpOutgoingResponse{
			NewStatus: desc.OperationStatus_Declined,
		}, err
	}

	err = u.commonRepo.Transactional(ctx, func(tx pgx.Tx) error {
		sender := domainRequest.Transaction.SenderID
		amount := domainRequest.Transaction.Amount

		txErr := u.commonRepo.LockClient(ctx, tx, sender)
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
		txErr = u.clientRepo.UpsertTransaction(ctx, tx, repository.MapSbpOutgoingDomainToTransaction(*domainRequest))
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return &desc.SbpOutgoingResponse{
			NewStatus: desc.OperationStatus_Declined,
		}, err
	}

	u.sbpIntegration.ToAnotherBank(ctx, *domainRequest.Transaction)

	return &desc.SbpOutgoingResponse{
		NewStatus: desc.OperationStatus_Approved,
	}, nil

}
