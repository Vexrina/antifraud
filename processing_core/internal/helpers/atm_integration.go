package helpers

import (
	"context"
	"time"

	"processing_core/internal/app/model"
)

type atmIntegration struct{}

func NewAtmIntegration() *atmIntegration {
	return &atmIntegration{}
}

func (integration *atmIntegration) GiveMoney(ctx context.Context, operation model.Transaction) {
	time.Sleep(10 * time.Millisecond)
}
