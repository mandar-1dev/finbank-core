-- CoreBank schema
-- Educational simulation only. No real money, no real banks, no auth.

CREATE DATABASE IF NOT EXISTS corebank_db
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE corebank_db;

SET NAMES utf8mb4;

-- ---------------------------------------------------------------------
-- customers
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS customers (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  customer_number VARCHAR(20)  NOT NULL UNIQUE,
  full_name       VARCHAR(120) NOT NULL,
  email           VARCHAR(160) NOT NULL UNIQUE,
  phone           VARCHAR(20)  NOT NULL,
  date_of_birth   DATE         NOT NULL,
  address         VARCHAR(255) NOT NULL,
  status          ENUM('ACTIVE','SUSPENDED') NOT NULL DEFAULT 'ACTIVE',
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- account_types
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS account_types (
  id          TINYINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code        VARCHAR(20) NOT NULL UNIQUE,   -- SAVINGS / CURRENT
  label       VARCHAR(40) NOT NULL,
  interest_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- accounts
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS accounts (
  id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  account_number     VARCHAR(20) NOT NULL UNIQUE,
  customer_id        BIGINT UNSIGNED NOT NULL,
  account_type_id    TINYINT UNSIGNED NOT NULL,
  balance            DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  available_balance  DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status             ENUM('ACTIVE','FROZEN','CLOSED') NOT NULL DEFAULT 'ACTIVE',
  opened_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_accounts_customer FOREIGN KEY (customer_id) REFERENCES customers(id),
  CONSTRAINT fk_accounts_type FOREIGN KEY (account_type_id) REFERENCES account_types(id),
  CONSTRAINT chk_balance_non_negative CHECK (balance >= 0)
) ENGINE=InnoDB;

CREATE INDEX idx_accounts_customer ON accounts(customer_id);

-- ---------------------------------------------------------------------
-- transactions
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS transactions (
  id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  transaction_ref  VARCHAR(40) NOT NULL UNIQUE,
  type             ENUM('DEPOSIT','WITHDRAWAL','TRANSFER','CARD_PAYMENT','LOAN_DISBURSEMENT','LOAN_REPAYMENT') NOT NULL,
  status           ENUM('PENDING','SUCCESS','FAILED','REVERSED') NOT NULL DEFAULT 'PENDING',
  source_account_id      BIGINT UNSIGNED NULL,
  destination_account_id BIGINT UNSIGNED NULL,
  amount           DECIMAL(18,2) NOT NULL,
  description      VARCHAR(255) NULL,
  idempotency_key  VARCHAR(80) NULL UNIQUE,
  risk_level       ENUM('LOW','MEDIUM','HIGH') NOT NULL DEFAULT 'LOW',
  created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_txn_source FOREIGN KEY (source_account_id) REFERENCES accounts(id),
  CONSTRAINT fk_txn_dest FOREIGN KEY (destination_account_id) REFERENCES accounts(id)
) ENGINE=InnoDB;

CREATE INDEX idx_txn_source ON transactions(source_account_id);
CREATE INDEX idx_txn_dest ON transactions(destination_account_id);
CREATE INDEX idx_txn_created ON transactions(created_at);

-- ---------------------------------------------------------------------
-- ledger_entries (double-entry bookkeeping)
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ledger_entries (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  transaction_id BIGINT UNSIGNED NOT NULL,
  account_id     BIGINT UNSIGNED NOT NULL,
  entry_type     ENUM('DEBIT','CREDIT') NOT NULL,
  amount         DECIMAL(18,2) NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_ledger_txn FOREIGN KEY (transaction_id) REFERENCES transactions(id),
  CONSTRAINT fk_ledger_account FOREIGN KEY (account_id) REFERENCES accounts(id)
) ENGINE=InnoDB;

CREATE INDEX idx_ledger_account ON ledger_entries(account_id);

-- ---------------------------------------------------------------------
-- beneficiaries
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS beneficiaries (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  customer_id     BIGINT UNSIGNED NOT NULL,
  name            VARCHAR(120) NOT NULL,
  account_number  VARCHAR(20) NOT NULL,
  ifsc_code       VARCHAR(15) NOT NULL DEFAULT 'CORE0000001',
  nickname        VARCHAR(60) NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_beneficiary_customer FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- cards
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cards (
  id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  account_id       BIGINT UNSIGNED NOT NULL,
  card_number      VARCHAR(20) NOT NULL UNIQUE,  -- fictional, e.g. 4582********9214
  card_holder      VARCHAR(120) NOT NULL,
  expiry_month     TINYINT UNSIGNED NOT NULL,
  expiry_year      SMALLINT UNSIGNED NOT NULL,
  status           ENUM('ACTIVE','FROZEN','CLOSED') NOT NULL DEFAULT 'ACTIVE',
  spending_limit   DECIMAL(18,2) NOT NULL DEFAULT 50000.00,
  created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_card_account FOREIGN KEY (account_id) REFERENCES accounts(id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- loans
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS loans (
  id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  customer_id      BIGINT UNSIGNED NOT NULL,
  account_id       BIGINT UNSIGNED NOT NULL,
  principal_amount DECIMAL(18,2) NOT NULL,
  interest_rate    DECIMAL(5,2) NOT NULL,
  tenure_months    SMALLINT UNSIGNED NOT NULL,
  emi_amount       DECIMAL(18,2) NOT NULL,
  status           ENUM('PENDING','APPROVED','REJECTED','ACTIVE','CLOSED') NOT NULL DEFAULT 'PENDING',
  applied_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_loan_customer FOREIGN KEY (customer_id) REFERENCES customers(id),
  CONSTRAINT fk_loan_account FOREIGN KEY (account_id) REFERENCES accounts(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS loan_payments (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  loan_id     BIGINT UNSIGNED NOT NULL,
  amount      DECIMAL(18,2) NOT NULL,
  paid_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_payment_loan FOREIGN KEY (loan_id) REFERENCES loans(id)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------------
-- audit_logs
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_logs (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  action       VARCHAR(60) NOT NULL,
  actor_type   ENUM('CUSTOMER','BANKER','SYSTEM') NOT NULL,
  actor_id     BIGINT UNSIGNED NULL,          -- customer id, when actor_type = CUSTOMER
  entity       VARCHAR(40) NOT NULL,          -- ACCOUNT / TRANSACTION / CARD / LOAN / BENEFICIARY
  entity_id    BIGINT UNSIGNED NULL,
  description  VARCHAR(255) NOT NULL,
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE INDEX idx_audit_created ON audit_logs(created_at);
