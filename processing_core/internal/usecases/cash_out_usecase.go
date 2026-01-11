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

//go:generate mockgen -source=cash_out_usecase.go -destination=./mocks/cash_out_usecase_mock.go -package=mocks
type (
	CashOutClientRepo interface {
		GetCurrentBalanceTx(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) (int64, error)
		UpdateBalance(ctx context.Context, tx pgx.Tx, clientID uuid.UUID, newBalance int64) error

		UpsertTransaction(ctx context.Context, tx pgx.Tx, transaction dbModel.TransactionsHistory) error
	}

	AntifraudCashOutCheck interface {
		CashOutCheck(ctx context.Context, operation *model.CashOutDomainRequest) error
	}

	CashOutInterface interface {
		// считаем что это асинхронщина, которая при падении в банкомате не даст списать деньги и если что их вернет на место физически.
		GiveMoney(ctx context.Context, operation model.Transaction)
	}

	cashOutUsecase struct {
		commonRepo CommonRepo
		clientRepo CashOutClientRepo

		antifraud AntifraudCashOutCheck
		atm       CashOutInterface
	}
)

func NewCashOutUsecase(
	commonRepo CommonRepo,
	clientRepo CashOutClientRepo,
	antifraud AntifraudCashOutCheck,
	atm CashOutInterface,
) *cashOutUsecase {
	return &cashOutUsecase{
		commonRepo: commonRepo,
		clientRepo: clientRepo,
		antifraud:  antifraud,
		atm:        atm,
	}
}

func (u *cashOutUsecase) Process(ctx context.Context, domainRequest *model.CashOutDomainRequest) (*desc.CashOutResponse, error) {
	afCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := u.antifraud.CashOutCheck(afCtx, domainRequest)
	// пришла ошибка из антифрода и это НЕ отмена контекста
	if err != nil && !(errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
		return &desc.CashOutResponse{
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

		txErr = u.clientRepo.UpsertTransaction(
			ctx,
			tx,
			repository.MapCashOutDomainToTransaction(*domainRequest),
		)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return &desc.CashOutResponse{
			NewStatus: desc.OperationStatus_Declined,
		}, err
	}

	u.atm.GiveMoney(ctx, *domainRequest.Transaction)

	return &desc.CashOutResponse{
		NewStatus: desc.OperationStatus_Approved,
	}, nil

}
