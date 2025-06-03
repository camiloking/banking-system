package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	// ARRANGE
	store := NewAccountStore()
	accountID := randomAccountID()
	initialBalance := float64(1000)
	timestamp := 1

	// ACT
	account := store.CreateAccount(timestamp, accountID, initialBalance)

	// ASSERT
	assert.NotNil(t, account, "expected account to be created")
	assert.Equal(t, accountID, account.accountID, "accountID mismatch")
	assert.Equal(t, initialBalance, account.balance, "balance mismatch")
	assert.Equal(t, timestamp, account.updatedAt, "updatedAt mismatch")
}

func TestTransfer(t *testing.T) {
	store := NewAccountStore()

	t.Run("Successful Transfer", func(t *testing.T) {
		// ARRANGE
		fromID := randomAccountID()
		toID := randomAccountID()
		initialBalance := float64(1000)
		transferAmount := float64(200)
		timestamp := 1

		store.CreateAccount(timestamp, fromID, initialBalance)
		store.CreateAccount(timestamp, toID, initialBalance)

		// ACT
		success, err := store.Transfer(timestamp+1, fromID, toID, transferAmount)

		// ASSERT
		assert.NoError(t, err, "unexpected error during transfer")
		assert.True(t, success, "expected transfer to succeed")

		fromAccount := store.accounts[fromID]
		toAccount := store.accounts[toID]

		assert.Equal(t, initialBalance-transferAmount, fromAccount.balance, "fromAccount balance mismatch")
		assert.Equal(t, initialBalance+transferAmount, toAccount.balance, "toAccount balance mismatch")
		assert.Equal(t, timestamp+1, fromAccount.updatedAt, "fromAccount updatedAt mismatch")
		assert.Equal(t, timestamp+1, toAccount.updatedAt, "toAccount updatedAt mismatch")
	})

	t.Run("Insufficient Balance", func(t *testing.T) {
		// ARRANGE
		fromID := randomAccountID()
		toID := randomAccountID()
		initialBalance := float64(100)
		transferAmount := float64(200)
		timestamp := 1

		store.CreateAccount(timestamp, fromID, initialBalance)
		store.CreateAccount(timestamp, toID, initialBalance)

		// ACT
		success, err := store.Transfer(timestamp+1, fromID, toID, transferAmount)

		// ASSERT
		assert.Error(t, err, "expected error due to insufficient balance")
		assert.False(t, success, "expected transfer to fail")

		fromAccount := store.accounts[fromID]
		toAccount := store.accounts[toID]

		assert.Equal(t, initialBalance, fromAccount.balance, "fromAccount balance mismatch")
		assert.Equal(t, initialBalance, toAccount.balance, "toAccount balance mismatch")
	})

	t.Run("Non-Existent Account", func(t *testing.T) {
		// ARRANGE
		fromID := randomAccountID()
		toID := "nonexistent"
		initialBalance := float64(1000)
		transferAmount := float64(200)
		timestamp := 1

		store.CreateAccount(timestamp, fromID, initialBalance)

		// ACT
		success, err := store.Transfer(timestamp+1, fromID, toID, transferAmount)

		// ASSERT
		assert.Error(t, err, "expected error due to non-existent account")
		assert.False(t, success, "expected transfer to fail")
	})
}

func TestSchedulePayment(t *testing.T) {
	store := NewAccountStore()

	t.Run("Successful Payment", func(t *testing.T) {
		// ARRANGE
		accountID := randomAccountID()
		initialBalance := float64(1000)
		paymentAmount := float64(200)
		delay := 1
		timestamp := int(time.Now().Unix())

		store.CreateAccount(timestamp, accountID, initialBalance)

		// ACT
		paymentID, err := store.SchedulePayment(timestamp, accountID, paymentAmount, delay)

		// ASSERT
		assert.NoError(t, err, "unexpected error during schedule payment")
		assert.NotNil(t, paymentID, "expected payment ID to be generated")
		assert.Equal(t, fmt.Sprintf("payment-%s-%d", accountID, 1), *paymentID, "expected one scheduled payment")

		// Wait for the payment to execute
		time.Sleep(time.Duration(delay+1) * time.Second)

		account := store.accounts[accountID]
		assert.Equal(t, initialBalance-paymentAmount, account.balance, "account balance mismatch after payment")
		assert.Equal(t, paymentAmount, account.totalTransferred, "total transferred mismatch after payment")
	})

	t.Run("Insufficient Balance", func(t *testing.T) {
		// ARRANGE
		accountID := randomAccountID()
		initialBalance := float64(100)
		paymentAmount := float64(200)
		delay := 1
		timestamp := int(time.Now().Unix())
		store.CreateAccount(timestamp, accountID, initialBalance)

		// ACT
		paymentID, err := store.SchedulePayment(timestamp, accountID, paymentAmount, delay)

		// ASSERT
		assert.NoError(t, err, "unexpected error during schedule payment")
		assert.NotNil(t, paymentID, "expected payment ID to be generated")

		// Wait for the payment to execute
		time.Sleep(time.Duration(delay+1) * time.Second)

		account := store.accounts[accountID]
		assert.Equal(t, initialBalance, account.balance, "account balance should remain unchanged due to insufficient funds")
		assert.Equal(t, float64(0), account.totalTransferred, "total transferred should remain unchanged")
	})

	t.Run("Non-Existent Account", func(t *testing.T) {
		// ARRANGE
		accountID := "nonexistent"
		paymentAmount := float64(200)
		delay := 1
		timestamp := int(time.Now().Unix())

		// ACT
		paymentID, err := store.SchedulePayment(timestamp, accountID, paymentAmount, delay)

		// ASSERT
		assert.Nil(t, paymentID, "expected no payment ID to be generated")
		assert.Error(t, err, "expected error due to non-existent account")
	})
}

func TestCancelScheduledPayment(t *testing.T) {
	store := NewAccountStore()

	t.Run("Successful Cancellation", func(t *testing.T) {
		// ARRANGE
		accountID := randomAccountID()
		initialBalance := float64(1000)
		paymentAmount := float64(200)
		delay := 2
		timestamp := int(time.Now().Unix())

		store.CreateAccount(timestamp, accountID, initialBalance)
		paymentID, err := store.SchedulePayment(timestamp, accountID, paymentAmount, delay)
		assert.NoError(t, err, "unexpected error during schedule payment")
		assert.NotNil(t, paymentID, "expected payment ID to be generated")

		// ACT
		err = store.CancelScheduledPayment(*paymentID)

		// ASSERT
		assert.NoError(t, err, "unexpected error during cancellation")
		_, exists := store.scheduledPayments[*paymentID]
		assert.False(t, exists, "payment should be removed from scheduled payments")
		account := store.accounts[accountID]
		assert.Equal(t, initialBalance, account.balance, "account balance mismatch")
	})

	t.Run("Non-Existent Payment", func(t *testing.T) {
		// ARRANGE
		nonExistentPaymentID := "nonexistent-payment"

		// ACT
		err := store.CancelScheduledPayment(nonExistentPaymentID)

		// ASSERT
		assert.Error(t, err, "expected error for non-existent payment")
		assert.Equal(t, "payment not found", err.Error(), "unexpected error message")
	})

	t.Run("Already Executed Payment", func(t *testing.T) {
		// ARRANGE
		accountID := randomAccountID()
		initialBalance := float64(1000)
		paymentAmount := float64(200)
		delay := 1
		timestamp := int(time.Now().Unix())

		store.CreateAccount(timestamp, accountID, initialBalance)
		paymentID, err := store.SchedulePayment(timestamp, accountID, paymentAmount, delay)
		assert.NoError(t, err, "unexpected error during schedule payment")
		assert.NotNil(t, paymentID, "expected payment ID to be generated")

		// Wait for the payment to execute
		time.Sleep(time.Duration(delay+1) * time.Second)

		// ACT
		err = store.CancelScheduledPayment(*paymentID)

		// ASSERT
		assert.Error(t, err, "expected error for already executed payment")
		assert.Equal(t, "payment already executed or cancelled", err.Error(), "unexpected error message")
	})
}
func TestMergeAccounts(t *testing.T) {
	store := NewAccountStore()

	t.Run("Successful Merge", func(t *testing.T) {
		// ARRANGE
		fromID := randomAccountID()
		toID := randomAccountID()
		fromInitialBalance := float64(500)
		toInitialBalance := float64(1000)
		fromTotalTransferred := float64(200)
		timestamp := 1

		fromAccount := store.CreateAccount(timestamp, fromID, fromInitialBalance)
		toAccount := store.CreateAccount(timestamp, toID, toInitialBalance)

		// Simulate some transfers for the "from" account
		fromAccount.totalTransferred = fromTotalTransferred

		// ACT
		err := store.MergeAccounts(timestamp+1, fromID, toID)

		// ASSERT
		assert.NoError(t, err, "unexpected error during merge")
		_, fromExists := store.accounts[fromID]
		assert.False(t, fromExists, "from account should be deleted after merge")

		mergedAccount := store.accounts[toID]
		assert.Equal(t, fromInitialBalance+toInitialBalance, mergedAccount.balance, "merged account balance mismatch")
		assert.Equal(t, toAccount.totalTransferred, mergedAccount.totalTransferred, "merged account total transferred mismatch")
		assert.Equal(t, timestamp+1, mergedAccount.updatedAt, "merged account updatedAt mismatch")
	})

	t.Run("Non-Existent From Account", func(t *testing.T) {
		// ARRANGE
		fromID := "nonexistent"
		toID := randomAccountID()
		toInitialBalance := float64(1000)
		timestamp := 1

		store.CreateAccount(timestamp, toID, toInitialBalance)

		// ACT
		err := store.MergeAccounts(timestamp+1, fromID, toID)

		// ASSERT
		assert.Error(t, err, "expected error for non-existent from account")
		assert.Equal(t, "one or both accounts do not exist", err.Error(), "unexpected error message")
	})

	t.Run("Non-Existent To Account", func(t *testing.T) {
		// ARRANGE
		fromID := randomAccountID()
		toID := "nonexistent"
		fromInitialBalance := float64(500)
		timestamp := 1

		store.CreateAccount(timestamp, fromID, fromInitialBalance)

		// ACT
		err := store.MergeAccounts(timestamp+1, fromID, toID)

		// ASSERT
		assert.Error(t, err, "expected error for non-existent to account")
		assert.Equal(t, "one or both accounts do not exist", err.Error(), "unexpected error message")
	})
}
func randomAccountID() string {
	newUUID := uuid.New()
	return newUUID.String()
}
