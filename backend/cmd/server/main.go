// Command server boots the CoreBank API: a local, educational banking
// simulation. It never touches a real bank, a real payment gateway, or
// real money — every customer, account, and balance is fictional and
// lives only in the local corebank_db MySQL database.
package main

import (
	"log"
	"net/http"

	"corebank/backend/internal/config"
	"corebank/backend/internal/db"
	"corebank/backend/internal/handlers"
	"corebank/backend/internal/middleware"
	"corebank/backend/internal/services"
)

func main() {
	cfg := config.Load()

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()
	log.Println("connected to MySQL at", cfg.DBHost+":"+cfg.DBPort, "db="+cfg.DBName)

	// --- services -------------------------------------------------------
	auditService := services.NewAuditLogService(conn)
	customerService := services.NewCustomerService(conn)
	accountService := services.NewAccountService(conn)
	transactionService := services.NewTransactionService(conn, auditService)
	transferService := services.NewTransferService(conn, auditService)
	beneficiaryService := services.NewBeneficiaryService(conn, auditService)
	cardService := services.NewCardService(conn, auditService)
	loanService := services.NewLoanService(conn, auditService)
	bankerService := services.NewBankerService(conn, auditService, accountService)

	// --- handlers ---------------------------------------------------------
	customerHandler := handlers.NewCustomerHandler(customerService, accountService, transactionService, beneficiaryService, cardService, loanService)
	accountHandler := handlers.NewAccountHandler(accountService, transactionService)
	transactionHandler := handlers.NewTransactionHandler(transactionService, transferService)
	beneficiaryHandler := handlers.NewBeneficiaryHandler(beneficiaryService)
	cardHandler := handlers.NewCardHandler(cardService)
	loanHandler := handlers.NewLoanHandler(loanService)
	bankerHandler := handlers.NewBankerHandler(bankerService, customerService, accountService, transactionService, cardService, loanService, auditService)

	mux := http.NewServeMux()

	// health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"corebank-api"}`))
	})

	// customers
	mux.HandleFunc("GET /api/customers", customerHandler.List)
	mux.HandleFunc("GET /api/customers/{id}", customerHandler.Get)
	mux.HandleFunc("GET /api/customers/{id}/accounts", customerHandler.Accounts)
	mux.HandleFunc("GET /api/customers/{id}/transactions", customerHandler.Transactions)
	mux.HandleFunc("GET /api/customers/{id}/beneficiaries", customerHandler.Beneficiaries)
	mux.HandleFunc("GET /api/customers/{id}/cards", customerHandler.Cards)
	mux.HandleFunc("GET /api/customers/{id}/loans", customerHandler.Loans)

	// accounts
	mux.HandleFunc("GET /api/accounts", accountHandler.List)
	mux.HandleFunc("GET /api/accounts/{id}", accountHandler.Get)
	mux.HandleFunc("GET /api/accounts/{id}/transactions", accountHandler.Transactions)

	// transactions
	mux.HandleFunc("POST /api/transactions/deposit", transactionHandler.Deposit)
	mux.HandleFunc("POST /api/transactions/withdraw", transactionHandler.Withdraw)
	mux.HandleFunc("POST /api/transactions/transfer", transactionHandler.Transfer)
	mux.HandleFunc("GET /api/transactions/{id}", transactionHandler.Get)

	// beneficiaries
	mux.HandleFunc("GET /api/beneficiaries/customer/{customerId}", beneficiaryHandler.ListByCustomer)
	mux.HandleFunc("POST /api/beneficiaries", beneficiaryHandler.Add)
	mux.HandleFunc("DELETE /api/beneficiaries/{id}", beneficiaryHandler.Remove)

	// cards
	mux.HandleFunc("GET /api/cards/customer/{customerId}", cardHandler.ListByCustomer)
	mux.HandleFunc("POST /api/cards/{id}/freeze", cardHandler.Freeze)
	mux.HandleFunc("POST /api/cards/{id}/unfreeze", cardHandler.Unfreeze)
	mux.HandleFunc("POST /api/cards/{id}/pay", cardHandler.Pay)

	// loans
	mux.HandleFunc("GET /api/loans/customer/{customerId}", loanHandler.ListByCustomer)
	mux.HandleFunc("POST /api/loans/estimate", loanHandler.Estimate)
	mux.HandleFunc("POST /api/loans/apply", loanHandler.Apply)

	// banker
	mux.HandleFunc("GET /api/banker/dashboard", bankerHandler.Dashboard)
	mux.HandleFunc("GET /api/banker/customers", bankerHandler.Customers)
	mux.HandleFunc("GET /api/banker/customers/{id}", bankerHandler.CustomerDetails)
	mux.HandleFunc("GET /api/banker/accounts", bankerHandler.Accounts)
	mux.HandleFunc("GET /api/banker/transactions", bankerHandler.Transactions)
	mux.HandleFunc("GET /api/banker/suspicious-transactions", bankerHandler.SuspiciousTransactions)
	mux.HandleFunc("GET /api/banker/cards", bankerHandler.Cards)
	mux.HandleFunc("GET /api/banker/loans", bankerHandler.Loans)
	mux.HandleFunc("GET /api/banker/audit-logs", bankerHandler.AuditLogs)
	mux.HandleFunc("POST /api/banker/accounts/{id}/freeze", bankerHandler.FreezeAccount)
	mux.HandleFunc("POST /api/banker/accounts/{id}/unfreeze", bankerHandler.UnfreezeAccount)

	handler := middleware.Recover(middleware.CORS(middleware.Logger(mux)))

	addr := ":" + cfg.ServerPort
	log.Println("CoreBank API listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
