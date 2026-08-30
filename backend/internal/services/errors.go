package services

import "corebank/backend/internal/util"

// These map directly to the exception types named in the project spec
// (CustomerNotFoundException, InsufficientBalanceException, ...). Go has
// no exceptions, so we express them as sentinel *util.APIError values
// returned up through the service → handler chain.

func errAccountNotFound() error     { return util.NotFound("Account") }
func errCustomerNotFound() error    { return util.NotFound("Customer") }
func errBeneficiaryNotFound() error { return util.NotFound("Beneficiary") }
func errCardNotFound() error        { return util.NotFound("Card") }
func errLoanNotFound() error        { return util.NotFound("Loan") }

func errAccountFrozen() error {
	return util.UnprocessableEntity("Account is frozen and cannot be used for this operation")
}

func errAccountClosed() error {
	return util.UnprocessableEntity("Account is closed")
}

func errInsufficientBalance() error {
	return util.UnprocessableEntity("Insufficient account balance")
}

func errInvalidAmount() error {
	return util.BadRequest("Amount must be greater than zero")
}

func errSameAccountTransfer() error {
	return util.BadRequest("Source and destination accounts must be different")
}

func errDuplicateTransaction() error {
	return util.Conflict("This transaction has already been processed")
}

func errSpendingLimitExceeded() error {
	return util.UnprocessableEntity("Card spending limit exceeded")
}
