package txmanager_test

import (
	// stdlib
	"context"
	"errors"
	"testing"
	"time"

	// internal
	"example.com/student-service/internal/txmanager"
	// external
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TxManagerTestSuite struct {
	suite.Suite
	ctx       context.Context
	container *postgres.PostgresContainer
	db        *sqlx.DB
	txMgr     *txmanager.Manager
}

func (s *TxManagerTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := postgres.Run(
		s.ctx,
		"docker.io/postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err)
	s.container = container

	dsn, err := container.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	db, err := sqlx.Open("postgres", dsn)
	s.Require().NoError(err)
	s.Require().NoError(waitForDB(db))

	s.db = db
	s.txMgr = txmanager.NewManager(db)
}

func (s *TxManagerTestSuite) TearDownSuite() {
	if err := s.db.Close(); err != nil {
		s.T().Logf("Error closing DB: %v", err)
	}
	err := s.container.Terminate(s.ctx)
	s.Require().NoError(err)
}

func (s *TxManagerTestSuite) TestManualSQL_Success() {
	_, err := s.db.ExecContext(s.ctx, `
        CREATE TABLE IF NOT EXISTS dummy (
            id TEXT PRIMARY KEY,
            value TEXT NOT NULL
        );
    `)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(s.ctx, `DELETE FROM dummy;`)
	s.Require().NoError(err)

	err = s.txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
		tx, err := txmanager.GetTx(ctx)
		s.Require().NoError(err)
		_, err = tx.ExecContext(ctx, `INSERT INTO dummy (id, value) VALUES ('id1', 'value1')`)
		return err
	})
	s.Require().NoError(err)

	var value string
	err = s.db.GetContext(s.ctx, &value, "SELECT value FROM dummy WHERE id = $1", "id1")
	s.Require().NoError(err)
	s.Equal("value1", value)
}

func (s *TxManagerTestSuite) TestParallelTransactions_NonConflicting() {
	_, err := s.db.ExecContext(s.ctx, `
        CREATE TABLE IF NOT EXISTS dummy (
            id TEXT PRIMARY KEY,
            value TEXT NOT NULL
        );
    `)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(s.ctx, `DELETE FROM dummy;`)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(s.ctx, `
        INSERT INTO dummy (id, value) VALUES ('id1', 'init'), ('id2', 'init');
    `)
	s.Require().NoError(err)

	ch1 := make(chan struct{})
	ch2 := make(chan struct{})

	go func() {
		err := s.txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
			tx, err := txmanager.GetTx(ctx)
			s.Require().NoError(err)

			_, err = tx.ExecContext(ctx, `UPDATE dummy SET value = 'tx1' WHERE id = 'id1'`)
			s.Require().NoError(err)

			ch1 <- struct{}{}
			<-ch2

			return nil
		})
		s.Require().NoError(err)
	}()

	go func() {
		err := s.txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
			tx, err := txmanager.GetTx(ctx)
			s.Require().NoError(err)

			<-ch1
			_, err = tx.ExecContext(ctx, `UPDATE dummy SET value = 'tx2' WHERE id = 'id2'`)
			s.Require().NoError(err)

			ch2 <- struct{}{}
			return nil
		})
		s.Require().NoError(err)
	}()

	time.Sleep(2 * time.Second)

	var v1, v2 string
	s.Require().NoError(s.db.Get(&v1, `SELECT value FROM dummy WHERE id = 'id1'`))
	s.Require().NoError(s.db.Get(&v2, `SELECT value FROM dummy WHERE id = 'id2'`))
	s.Equal("tx1", v1)
	s.Equal("tx2", v2)
}

func (s *TxManagerTestSuite) TestParallelTransactions_Conflicting() {
	_, err := s.db.ExecContext(s.ctx, `
        CREATE TABLE IF NOT EXISTS dummy (
            id TEXT PRIMARY KEY,
            value TEXT NOT NULL
        );
    `)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(s.ctx, `DELETE FROM dummy;`)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(s.ctx, `INSERT INTO dummy (id, value) VALUES ('id1', 'init');`)
	s.Require().NoError(err)

	chReady := make(chan struct{})
	chContinue := make(chan struct{})
	errCh := make(chan error, 2)

	// Tx1
	go func() {
		err := s.txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
			tx, err := txmanager.GetTx(ctx)
			s.Require().NoError(err)

			s.T().Log("Tx1: reading row")
			var val string
			err = tx.GetContext(ctx, &val, `SELECT value FROM dummy WHERE id = 'id1'`)
			s.Require().NoError(err)

			s.T().Log("Tx1: signaling Tx2 to start")
			chReady <- struct{}{}

			s.T().Log("Tx1: waiting for Tx2 to read and write")
			<-chContinue

			s.T().Log("Tx1: updating row")
			_, err = tx.ExecContext(ctx, `UPDATE dummy SET value = 'tx1' WHERE id = 'id1'`)
			if err != nil {
				return err
			}

			return nil
		})
		s.T().Logf("Tx1 done: %v", err)
		errCh <- err
	}()

	// Tx2
	go func() {
		err := s.txMgr.WithTransaction(s.ctx, func(ctx context.Context) error {
			tx, err := txmanager.GetTx(ctx)
			s.Require().NoError(err)

			s.T().Log("Tx2: waiting for Tx1 to read")
			<-chReady

			s.T().Log("Tx2: reading row")
			var val string
			err = tx.GetContext(ctx, &val, `SELECT value FROM dummy WHERE id = 'id1'`)
			s.Require().NoError(err)

			s.T().Log("Tx2: updating row")
			_, err = tx.ExecContext(ctx, `UPDATE dummy SET value = 'tx1' WHERE id = 'id1'`)
			if err != nil {
				return err
			}

			s.T().Log("Tx2: signaling Tx1 to continue")
			chContinue <- struct{}{}

			return nil
		})
		s.T().Logf("Tx2 done: %v", err)
		errCh <- err
	}()

	err1 := <-errCh
	err2 := <-errCh

	s.T().Logf("Tx1 error: %v", err1)
	s.T().Logf("Tx2 error: %v", err2)

	s.True(err1 == nil || err2 == nil, "one transaction should succeed")
	s.True(err1 != nil || err2 != nil, "one transaction should fail due to serialization conflict")
}

func TestTxManagerTestSuite(t *testing.T) {
	suite.Run(t, new(TxManagerTestSuite))
}

func waitForDB(db *sqlx.DB) error {
	const maxAttempts = 10
	const delay = time.Second
	for i := 0; i < maxAttempts; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return errors.New("database not reachable after waiting")
}
