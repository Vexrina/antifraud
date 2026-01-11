package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/inf.v0"

	"antifraud/internal/constants"
)

func TestSpent3H(t *testing.T) {
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
CREATE TABLE IF NOT EXISTS spent_3h (
	user_id uuid,
	window_start timestamp,
	total_spent decimal,
	PRIMARY KEY ((user_id), window_start)
);
`).Exec()
					require.NoError(t, err)

					total := inf.NewDec(98765, 0)
					err = session.Query(`
INSERT INTO spent_3h (user_id, window_start, total_spent)
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
				window: time.Now().UTC().Truncate(time.Hour),
			},
			want: 98765,
		},
		{
			name: "no rows",
			fields: fields{
				prepare: func(session *gocql.Session, a *args) {
					err := session.Query(`
CREATE TABLE IF NOT EXISTS spent_3h (
	user_id uuid,
	window_start timestamp,
	total_spent decimal,
	PRIMARY KEY ((user_id), window_start)
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
				window: time.Now().UTC().Truncate(time.Hour),
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := newTestSession(t)
			t.Cleanup(func() {
				session.Query(`DELETE FROM spent_3h WHERE user_id = ? AND window_start = ?`, tt.args.filter.UserID, tt.args.window).Exec()
				session.Query(`DROP TABLE IF EXISTS spent_3h`).Exec()
				session.Close()
			})

			fs := &CassandraFeatureStore{session: session}

			tt.fields.prepare(session, &tt.args)

			ctx := context.Background()
			res, err := fs.getSpent3H(ctx, tt.args.filter)
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

func TestGetPartnersLastNBuckets(t *testing.T) {
	type args struct {
		table    string
		filter   *constants.FeatureFilter
		window   time.Time
		partners []gocql.UUID
		n        int
	}
	type fields struct {
		prepare func(session *gocql.Session, a *args)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr string
	}{
		{
			name: "got rows - sbp",
			fields: fields{
				prepare: func(session *gocql.Session, a *args) {
					err := session.Query(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	user_id uuid,
	window_start timestamp,
	partner_id uuid,
	PRIMARY KEY (user_id, window_start, partner_id)
) WITH CLUSTERING ORDER BY (window_start DESC, partner_id ASC)
  AND default_time_to_live = 172800;
`, a.table)).Exec()
					require.NoError(t, err)

					for _, p := range a.partners {
						err = session.Query(fmt.Sprintf(`
INSERT INTO %s (user_id, window_start, partner_id)
VALUES (?, ?, ?) USING TTL 172800
`, a.table), a.filter.UserID, a.window, p).Exec()
						require.NoError(t, err)
					}
				},
			},
			args: args{
				table:  "sbp_partners_30m",
				filter: &constants.FeatureFilter{UserID: mustUUID(), Limit: 5},
				window: time.Now().UTC().Truncate(time.Hour),
				partners: []gocql.UUID{
					mustUUID(), mustUUID(), mustUUID(),
				},
				n: 5,
			},
			wantErr: "",
		},
		{
			name: "no rows - internal",
			fields: fields{
				prepare: func(session *gocql.Session, a *args) {
					err := session.Query(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	user_id uuid,
	window_start timestamp,
	partner_id uuid,
	PRIMARY KEY (user_id, window_start, partner_id)
) WITH CLUSTERING ORDER BY (window_start DESC, partner_id ASC)
  AND default_time_to_live = 172800;
`, a.table)).Exec()
					require.NoError(t, err)
				},
			},
			args: args{
				table:    "internal_partners_30m",
				filter:   &constants.FeatureFilter{UserID: mustUUID(), Limit: 5},
				window:   time.Now().UTC().Truncate(time.Hour),
				partners: nil,
				n:        5,
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := newTestSession(t)
			t.Cleanup(func() {
				for _, p := range tt.args.partners {
					session.Query(fmt.Sprintf(`DELETE FROM %s WHERE user_id = ? AND window_start = ? AND partner_id = ?`, tt.args.table),
						tt.args.filter.UserID, tt.args.window, p).Exec()
				}
				session.Query(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tt.args.table)).Exec()
				session.Close()
			})

			fs := &CassandraFeatureStore{session: session}
			tt.fields.prepare(session, &tt.args)

			ctx := context.Background()
			got, err := fs.getPartnersLastNBuckets(ctx, tt.args.table, tt.args.filter, tt.args.n)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			}
			require.NoError(t, err)

			// Преобразуем UUID в строки для сравнения
			var want []string
			for _, p := range tt.args.partners {
				want = append(want, p.String())
			}

			if len(want) == 0 {
				assert.Nil(t, got)
			} else {
				assert.ElementsMatch(t, want, got)
			}
		})
	}
}
