package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Account struct {
	accountID        string
	updatedAt        int
	balance          float64
	totalTransferred float64
}

type AccountStore struct {
	mu                sync.RWMutex
	accounts          map[string]*Account
	nextPaymentID     int
	scheduledPayments map[string]*time.Timer
}

func NewAccountStore() *AccountStore {
	return &AccountStore{
		accounts:          make(map[string]*Account),
		nextPaymentID:     1,
		scheduledPayments: make(map[string]*time.Timer),
	}
}

func (s *AccountStore) CreateAccount(timestamp int, accountID string, initialBalance float64) *Account {
	s.mu.Lock()
	defer s.mu.Unlock()

	account := &Account{
		accountID:        accountID,
		updatedAt:        timestamp,
		balance:          initialBalance,
		totalTransferred: 0,
	}
	s.accounts[accountID] = account
	return account
}

func (s *AccountStore) Transfer(timestamp int, fromID, toID string, amount float64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromAccount, fromExists := s.accounts[fromID]
	toAccount, toExists := s.accounts[toID]

	if !fromExists || !toExists {
		return false, errors.New("one or both accounts do not exist")
	}

	if fromAccount.balance < amount {
		return false, errors.New("insufficient balance in the from account")
	}

	fromAccount.balance -= amount
	fromAccount.totalTransferred += amount
	fromAccount.updatedAt = timestamp

	toAccount.balance += amount
	toAccount.updatedAt = timestamp

	return true, nil
}

// Level 3 - Schedule Payment (Completed in the assessment) and Cancel Payment
func (s *AccountStore) SchedulePayment(timestamp int, accountID string, amount float64, delaySeconds int) (*string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.accounts[accountID]
	if !exists {
		return nil, errors.New("account does not exist")
	}

	executeAt := time.Unix(int64(timestamp), 0).Add(time.Duration(delaySeconds) * time.Second)
	delayDuration := time.Until(executeAt)
	if delayDuration <= 0 {
		delayDuration = 0
	}
	timer := time.AfterFunc(delayDuration, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		acc, exists := s.accounts[accountID]
		if !exists {
			return
		}
		if acc.balance < amount {
			return
		}
		acc.balance -= amount
		acc.totalTransferred += amount
	})

	paymentID := fmt.Sprintf("payment-%s-%d", accountID, s.nextPaymentID)
	s.scheduledPayments[paymentID] = timer

	return &paymentID, nil
}

func (s *AccountStore) CancelScheduledPayment(paymentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	timer, exists := s.scheduledPayments[paymentID]
	if !exists {
		return errors.New("payment not found")
	}

	// Stop the timer if it is still running
	stopped := timer.Stop()
	if !stopped {
		return errors.New("payment already executed or cancelled")
	}

	// Remove the payment from the scheduled payments map
	delete(s.scheduledPayments, paymentID)
	return nil
}

// Level 4 - Merge Accounts
func (s *AccountStore) MergeAccounts(timestamp int, fromID, toID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromAccount, fromExists := s.accounts[fromID]
	toAccount, toExists := s.accounts[toID]

	if !fromExists || !toExists {
		return errors.New("one or both accounts do not exist")
	}

	toAccount.balance += fromAccount.balance
	toAccount.totalTransferred += fromAccount.totalTransferred
	toAccount.updatedAt = timestamp

	delete(s.accounts, fromID)
	return nil
}
