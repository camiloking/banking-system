Assessment Follow-Up – Banking System

This repository contains my implementations of the final sections of a recent Industry Coding Assessment.

## ✍️ Context

During the timed session, I wasn't able to complete all sections of the assessment. This project includes solutions to the **remaining tasks that I wasn’t able to submit during the assessment**, reconstructed from memory shortly afterward.

Specifically, it was the CancelScheduledPayment and MergeAccounts methods that I didn't have time to implement.

While the method signatures and return values may not exactly match the originals, I’ve tried preserving the core logic and intent of the problems.

## 🧠 What’s Included

- A self-contained solution for the remaining features of the banking system prompt.
- Tests for the new methods in `bank_account_system_test.go`.
- Clear separation of logic for ease of review and readability.

## ✅ Why This Exists

I wanted to demonstrate how I would have completed the assessment had I had more time. I believe the code here reflects my ability to:
- Understand and implement the requirements,
- Write clean and testable code,
- Reason about system behavior under evolving functionality.

## 🧪 Tests

### 1. Clone the repo
```
git clone https://github.com/camiloking/banking-system
cd banking-system
```

### 2. Initialize Go modules (if not already done)
`go mod init github.com/camiloking/banking-system`

### 3. Tidy up dependencies
`go mod tidy`

### 4. Run tests
`go test -v`
