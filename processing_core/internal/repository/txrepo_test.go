package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestCommonRepository_RollBackUnlessCommitted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := NewTestDB(ctx, t)

	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Errorf("can't begin transaction: %v", err)
	}

	db.RollBackUnlessCommitted(ctx, tx)

	err = db.CommitTx(ctx, tx)
	assert.True(t, errors.Is(err, pgx.ErrTxClosed))
}

func TestCommonRepository_Transactional(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := NewTestDB(ctx, t)

	err := db.Transactional(ctx, func(tx pgx.Tx) error {
		return nil
	})

	assert.Nil(t, err)
}

func TestCommonRepository_RollBackUnlessCommitted_NilTx(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := NewTestDB(ctx, t)

	tx := pgx.Tx(nil)

	defer func() {
		res := recover()
		if res != nil {
			t.Errorf("test ended with panic: %v", res)
		}
	}()

	db.RollBackUnlessCommitted(ctx, tx)
}

func TestCommonRepository_LockClient_SameClient(t *testing.T) {
	t.Parallel()
	// тестируем, что пока клиент заблокирован, он не может быть заблокирован еще раз
	ctx := context.Background()
	db := NewTestDB(ctx, t)

	tx1, err := db.BeginTx(ctx)
	if err != nil {
		t.Errorf("can't get first transaction: %v", err)
	}

	tx2, err := db.BeginTx(ctx)
	if err != nil {
		t.Errorf("can't get second transaction: %v", err)
	}

	clientID := uuid.New()

	err = db.LockClient(ctx, tx1, clientID)
	if err != nil {
		t.Errorf("can't lock client first time: %v", err)
	}

	// пытаемся еще раз заблокировать клиента во второй горутине
	durationCh := make(chan time.Duration)
	go func() {
		start := time.Now()
		// так как клиент уже заблокирован в tx1, эта транзакция будет ждать разблокировки клиента
		lockErr := db.LockClient(ctx, tx2, clientID)
		if lockErr != nil {
			t.Errorf("can't lock client second time: %v", err)
		}

		// замеряем время, в течение которого транзакция ждала разблокировки клиента
		duration := time.Since(start)
		durationCh <- duration
	}()

	// имитируем работу
	workTime := time.Second * 5
	time.Sleep(workTime)

	// разблокируем клиента
	err = tx1.Commit(ctx)
	if err != nil {
		t.Errorf("can't commit tx1: %v", err)
	}

	// получаем время, в течение которого клиент был заблокирован
	waitTime := <-durationCh

	err = tx2.Commit(ctx)
	if err != nil {
		t.Errorf("can't commit tx2: %v", err)
	}

	// проверяем, что вторая транзакция ждал все время, пока первая держал клиента заблокированным
	assert.True(t, waitTime >= workTime)
}

func TestCommonRepository_LockExternalKey_DifferentClients(t *testing.T) {
	t.Parallel()
	// тестируем, что блокировки двух разных клиентов не влияют друг на друга
	ctx := context.Background()
	db := NewTestDB(ctx, t)

	tx1, err := db.BeginTx(ctx)
	if err != nil {
		t.Errorf("can't get first transaction: %v", err)
	}

	tx2, err := db.BeginTx(ctx)
	if err != nil {
		t.Errorf("can't get second transaction: %v", err)
	}

	clientID1 := uuid.New()
	clientID2 := uuid.New()

	// блокируем первого клиента
	err = db.LockClient(ctx, tx1, clientID1)
	if err != nil {
		t.Errorf("can't lock client first time: %v", err)
	}

	// пытаемся заблокировать второго клиента
	durationCh := make(chan time.Duration)
	go func() {
		start := time.Now()
		// так как uuid клиентов различаются, блокировка должна произойти мгновенно
		lockErr := db.LockClient(ctx, tx2, clientID2)
		if lockErr != nil {
			t.Errorf("can't lock client second time: %v", err)
		}

		// замеряем время, в течение которого клиент был заблокирован
		duration := time.Since(start)
		durationCh <- duration
	}()

	// имитируем работу
	workTime := time.Second * 5
	time.Sleep(workTime)

	// разблокируем первого клиента
	err = tx1.Commit(ctx)
	if err != nil {
		t.Errorf("can't commit tx1: %v", err)
	}

	// получаем время, в течение которого клиент был заблокирован
	waitTime := <-durationCh

	// разблокируем второго клиента
	err = tx2.Commit(ctx)
	if err != nil {
		t.Errorf("can't commit tx2: %v", err)
	} // проверяем, что второй клиент был заблокирован раньше, чем разблокирован первый
	assert.True(t, waitTime < workTime)
}
