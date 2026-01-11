package helpers

import (
	"context"
	"time"

	"processing_core/internal/app/model"
)

type sbpIntegration struct{}

func NewSbpIntegration() *sbpIntegration {
	return &sbpIntegration{}
}

func (integration *sbpIntegration) ToAnotherBank(ctx context.Context, operation model.Transaction) {
	time.Sleep(10 * time.Millisecond)
}
