package repository

import (
	"context"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/inf.v0"

	"antifraud/internal/constants"
)

func TestCashOut30M(t *testing.T) {
	type args struct {
		filter *constants.FeatureFilter
		window time.Time
	}
	type fields struct {
		prepare func(session *gocql.Session, a *args)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr string
	}{
		{
			name: "got rows",
			fields: fields{
				prepare: func(session *gocql.Session, a *args) {
					err := session.Query(`
CREATE TABLE IF NOT EXISTS cash_out_30m (
	user_id uuid,
	window_start timestamp,
	total_cashout decimal,
	PRIMARY KEY (user_id, window_start)
);
`).Exec()
					require.NoError(t, err)

					total := inf.NewDec(12345, 0)
					err = session.Query(`
	INSERT INTO cash_out_30m (user_id, window_start, total_cashout)
	VALUES (?, ?, ?)
`, a.filter.UserID, a.window, total).Exec()
					require.NoError(t, err)
				},
			},
			args: args{
				filter: &constants.FeatureFilter{
					UserID: mustUUID(),
					Limit:  1,
				},
				window: time.Now().UTC().Truncate(30 * time.Minute),
			},
			want: 12345,
		},
		{
			name: "no rows",
			fields: fields{
				prepare: func(session *gocql.Session, a *args) {
					err := session.Query(`
CREATE TABLE IF NOT EXISTS cash_out_30m (
	user_id uuid,
	window_start timestamp,
	total_cashout decimal,
	PRIMARY KEY (user_id, window_start)
);
`).Exec()
					require.NoError(t, err)
				},
			},
			args: args{
				filter: &constants.FeatureFilter{
					UserID: mustUUID(),
					Limit:  1,
				},
				window: time.Now().UTC().Truncate(30 * time.Minute),
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := newTestSession(t)
			t.Cleanup(func() {
				session.Query(`DELETE FROM cash_out_30m WHERE user_id = ? AND window_start = ?`, tt.args.filter.UserID, tt.args.window).Exec()
				session.Query(`DROP TABLE IF EXISTS cash_out_30m`).Exec()
				session.Close()
			})

			fs := &CassandraFeatureStore{session: session}

			tt.fields.prepare(session, &tt.args)

			ctx := context.Background()
			res, err := fs.getCashOut30M(ctx, tt.args.filter)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, res[0])
		})
	}
}
