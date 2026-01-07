package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"antifraud/internal/app/model"
	"antifraud/internal/usecases/mocks"
)

func TestSbpOut_Check(t *testing.T) {
	type args struct {
		ctx         context.Context
		transaction *model.DomainTransaction
	}
	tests := []struct {
		name      string
		rules     func(ctrl *gomock.Controller) []Rule
		args      args
		wantErr   bool
		errSubstr string
	}{
		{
			name: "all rules ShouldRun false",
			rules: func(ctrl *gomock.Controller) []Rule {
				r1 := mocks.NewMockRule(ctrl)
				r1.EXPECT().ShouldRun(gomock.Any(), gomock.Any()).Return(false)
				// Check не должен вызываться
				return []Rule{r1}
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
			wantErr: false,
		},
		{
			name: "one rule returns error",
			rules: func(ctrl *gomock.Controller) []Rule {
				r1 := mocks.NewMockRule(ctrl)
				r1.EXPECT().ShouldRun(gomock.Any(), gomock.Any()).Return(true)
				r1.EXPECT().Check(gomock.Any(), gomock.Any()).Return(errors.New("declined"))
				return []Rule{r1}
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
			wantErr:   true,
			errSubstr: "declined",
		},
		{
			name: "multiple rules, one error",
			rules: func(ctrl *gomock.Controller) []Rule {
				r1 := mocks.NewMockRule(ctrl)
				r2 := mocks.NewMockRule(ctrl)
				r2.EXPECT().ShouldRun(gomock.Any(), gomock.Any()).Return(true)
				r2.EXPECT().Check(gomock.Any(), gomock.Any()).Return(errors.New("rule2 failed"))
				return []Rule{r1, r2}
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
			wantErr:   true,
			errSubstr: "rule2 failed",
		},
		{
			name: "all rules pass",
			rules: func(ctrl *gomock.Controller) []Rule {
				r1 := mocks.NewMockRule(ctrl)
				r2 := mocks.NewMockRule(ctrl)
				r1.EXPECT().ShouldRun(gomock.Any(), gomock.Any()).Return(true)
				r1.EXPECT().Check(gomock.Any(), gomock.Any()).Return(nil)
				r2.EXPECT().ShouldRun(gomock.Any(), gomock.Any()).Return(false)
				// Check для r2 не вызывается
				return []Rule{r1, r2}
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			c := &SbpOutCheck{
				&RuleChecker{rules: tt.rules(ctrl)},
			}
			err := c.Check(tt.args.ctx, tt.args.transaction)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("Check() error = %v, want substring %q", err, tt.errSubstr)
			}
		})
	}
}
