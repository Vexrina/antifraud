package helpers

import (
	"context"
	"fmt"
	"processing_core/internal/app/model"
	"processing_core/pkg/antifraud"
	"time"

	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	cashOutCheckTimeout  = 200 * time.Hour
	sbpOutCheckTimeout   = 250 * time.Millisecond
	internalCheckTimeout = 400 * time.Millisecond
)

type afIntegration struct {
	af antifraud.OnlineCheckClient
}

//go:generate mockgen -source=./../../pkg/antifraud/antifraud_grpc.pb.go -destination=mocks/antifraud_client_mock.go -package=mocks
func NewAfIntegration(af antifraud.OnlineCheckClient) *afIntegration {
	return &afIntegration{
		af: af,
	}
}

func (af *afIntegration) CashOutCheck(ctx context.Context, operation *model.CashOutDomainRequest) error {
	ctx = context.WithoutCancel(ctx)
	//ctx, cancel := context.WithTimeout(ctx, cashOutCheckTimeout)
	//defer cancel()
	res, err := af.af.CashOut(ctx, mapCashOutToAntifraudTransaction(operation))
	if err != nil {
		// todo log
		return nil
	}
	if res.NewStatus == antifraud.OperationStatus_Declined {
		return fmt.Errorf("operation declined")
	}
	return nil
}

func mapCashOutToAntifraudTransaction(domain *model.CashOutDomainRequest) *antifraud.Transaction {
	return &antifraud.Transaction{
		Id:            domain.Transaction.ID,
		TransactionId: domain.Transaction.TransactionID.String(),
		CreatedAt:     timestamppb.New(*domain.Transaction.CreatedAt),
		Amount:        domain.Transaction.Amount,
		Currency:      fmt.Sprint(domain.Transaction.Currency),
		Merchant:      domain.Transaction.Merchant,
		Country:       domain.Transaction.Country,
		SenderId:      domain.Transaction.SenderID.String(),
		ReceiverId:    nil,
		Bic:           nil,
		AtmId:         lo.ToPtr(domain.AtmID.String()),
	}
}

func (af *afIntegration) InternalCheck(ctx context.Context, operation *model.InternalDomainRequest) error {
	ctx, cancel := context.WithTimeout(ctx, internalCheckTimeout)
	defer cancel()
	res, err := af.af.Internal(ctx, mapInternalToAntifraudTransaction(operation))
	if err != nil {
		// todo log
		return nil
	}
	if res.NewStatus == antifraud.OperationStatus_Declined {
		return fmt.Errorf("operation declined")
	}
	return nil
}

func mapInternalToAntifraudTransaction(domain *model.InternalDomainRequest) *antifraud.Transaction {
	return &antifraud.Transaction{
		Id:            domain.Transaction.ID,
		TransactionId: domain.Transaction.TransactionID.String(),
		CreatedAt:     timestamppb.New(*domain.Transaction.CreatedAt),
		Amount:        domain.Transaction.Amount,
		Currency:      fmt.Sprint(domain.Transaction.Currency),
		Merchant:      domain.Transaction.Merchant,
		Country:       domain.Transaction.Country,
		SenderId:      domain.Transaction.SenderID.String(),
		ReceiverId:    lo.ToPtr(domain.ReceiverId.String()),
		Bic:           nil,
		AtmId:         nil,
	}
}

func (af *afIntegration) SbpOutgoingCheck(ctx context.Context, operation *model.SbpOutgoingDomainRequest) error {
	ctx, cancel := context.WithTimeout(ctx, sbpOutCheckTimeout)
	defer cancel()
	res, err := af.af.SbpOutgoing(ctx, mapSbpOutgoingToAntifraudTransaction(operation))
	if err != nil {
		// todo log
		return nil
	}
	if res.NewStatus == antifraud.OperationStatus_Declined {
		return fmt.Errorf("operation declined")
	}
	return nil
}

func mapSbpOutgoingToAntifraudTransaction(domain *model.SbpOutgoingDomainRequest) *antifraud.Transaction {
	return &antifraud.Transaction{
		Id:            domain.Transaction.ID,
		TransactionId: domain.Transaction.TransactionID.String(),
		CreatedAt:     timestamppb.New(*domain.Transaction.CreatedAt),
		Amount:        domain.Transaction.Amount,
		Currency:      fmt.Sprint(domain.Transaction.Currency),
		Merchant:      domain.Transaction.Merchant,
		Country:       domain.Transaction.Country,
		SenderId:      domain.Transaction.SenderID.String(),
		ReceiverId:    lo.ToPtr(domain.ReceiverId.String()),
		Bic:           lo.ToPtr(domain.ReceiverId.String()),
		AtmId:         nil,
	}
}
