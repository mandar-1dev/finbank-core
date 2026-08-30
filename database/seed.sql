USE corebank_db;

-- ---------------------------------------------------------------------
-- account types
-- ---------------------------------------------------------------------
INSERT INTO account_types (code, label, interest_rate) VALUES
  ('SAVINGS', 'Savings Account', 3.50),
  ('CURRENT', 'Current Account', 0.00)
ON DUPLICATE KEY UPDATE label = VALUES(label);

-- ---------------------------------------------------------------------
-- customers
-- ---------------------------------------------------------------------
INSERT INTO customers (customer_number, full_name, email, phone, date_of_birth, address, status) VALUES
  ('CUST1001', 'Rahul Sharma',    'rahul.sharma@example.com',    '9820011122', '1996-04-12', '221 MG Road, Pune',        'ACTIVE'),
  ('CUST1002', 'Priya Mehta',     'priya.mehta@example.com',     '9820011123', '1994-09-03', '48 FC Road, Pune',         'ACTIVE'),
  ('CUST1003', 'Arjun Patil',     'arjun.patil@example.com',     '9820011124', '1990-01-21', '12 Koregaon Park, Pune',   'ACTIVE'),
  ('CUST1004', 'Sneha Kulkarni',  'sneha.kulkarni@example.com',  '9820011125', '1998-07-17', '76 Baner Road, Pune',      'ACTIVE'),
  ('CUST1005', 'Aditya Deshmukh', 'aditya.deshmukh@example.com', '9820011126', '1992-11-05', '9 Aundh, Pune',            'ACTIVE'),
  ('CUST1006', 'Neha Joshi',      'neha.joshi@example.com',      '9820011127', '1995-03-29', '33 Kothrud, Pune',         'ACTIVE');

-- ---------------------------------------------------------------------
-- accounts (account_type_id 1 = SAVINGS, 2 = CURRENT)
-- ---------------------------------------------------------------------
INSERT INTO accounts (account_number, customer_id, account_type_id, balance, available_balance, status) VALUES
  ('4821000000004821', 1, 1, 85000.00,  85000.00, 'ACTIVE'),   -- Rahul savings
  ('9034000000009034', 1, 2, 160000.00, 160000.00, 'ACTIVE'),  -- Rahul current
  ('7312000000007312', 2, 1, 125000.00, 125000.00, 'ACTIVE'),  -- Priya savings
  ('1945000000001945', 3, 2, 250000.00, 250000.00, 'ACTIVE'),  -- Arjun current
  ('5567000000005567', 4, 1, 42000.00,  42000.00, 'ACTIVE'),   -- Sneha savings
  ('6678000000006678', 5, 1, 98000.00,  98000.00, 'ACTIVE'),   -- Aditya savings
  ('8890000000008890', 6, 2, 310000.00, 310000.00, 'FROZEN');  -- Neha current (frozen, for demo)

-- ---------------------------------------------------------------------
-- beneficiaries
-- ---------------------------------------------------------------------
INSERT INTO beneficiaries (customer_id, name, account_number, ifsc_code, nickname) VALUES
  (1, 'Priya Mehta',  '7312000000007312', 'CORE0000001', 'Priya'),
  (1, 'Arjun Patil',  '1945000000001945', 'CORE0000001', 'Arjun'),
  (2, 'Rahul Sharma', '4821000000004821', 'CORE0000001', 'Rahul'),
  (3, 'Sneha Kulkarni','5567000000005567','CORE0000001', 'Sneha');

-- ---------------------------------------------------------------------
-- cards
-- ---------------------------------------------------------------------
INSERT INTO cards (account_id, card_number, card_holder, expiry_month, expiry_year, status, spending_limit) VALUES
  (1, '4582000000009214', 'RAHUL SHARMA',    8, 2029, 'ACTIVE', 50000.00),
  (3, '4582000000004417', 'PRIYA MEHTA',     3, 2030, 'ACTIVE', 75000.00),
  (4, '4582000000001120', 'ARJUN PATIL',     11, 2028, 'ACTIVE', 100000.00);

-- ---------------------------------------------------------------------
-- loans
-- ---------------------------------------------------------------------
INSERT INTO loans (customer_id, account_id, principal_amount, interest_rate, tenure_months, emi_amount, status) VALUES
  (1, 1, 200000.00, 10.50, 24, 9273.00, 'ACTIVE'),
  (3, 4, 500000.00, 9.00,  36, 15900.00, 'PENDING');

-- ---------------------------------------------------------------------
-- sample transactions + ledger entries
-- ---------------------------------------------------------------------
-- Salary deposit into Rahul's savings account
INSERT INTO transactions (transaction_ref, type, status, source_account_id, destination_account_id, amount, description, risk_level)
VALUES ('TXN-SEED-0001', 'DEPOSIT', 'SUCCESS', NULL, 1, 65000.00, 'Salary simulation', 'LOW');
INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount) VALUES (1, 1, 'CREDIT', 65000.00);

-- Card payment simulation on Rahul's account
INSERT INTO transactions (transaction_ref, type, status, source_account_id, destination_account_id, amount, description, risk_level)
VALUES ('TXN-SEED-0002', 'CARD_PAYMENT', 'SUCCESS', 1, NULL, 2400.00, 'Amazon Simulation', 'LOW');
INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount) VALUES (2, 1, 'DEBIT', 2400.00);

-- Transfer: Rahul -> Priya
INSERT INTO transactions (transaction_ref, type, status, source_account_id, destination_account_id, amount, description, risk_level, idempotency_key)
VALUES ('TXN-SEED-0003', 'TRANSFER', 'SUCCESS', 1, 3, 10000.00, 'Transfer to Priya', 'LOW', 'seed-transfer-0003');
INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount) VALUES (3, 1, 'DEBIT', 10000.00), (3, 3, 'CREDIT', 10000.00);

-- audit logs for the seeded actions
INSERT INTO audit_logs (action, actor_type, actor_id, entity, entity_id, description) VALUES
  ('DEPOSIT', 'SYSTEM', NULL, 'TRANSACTION', 1, 'Seed salary deposit into account ****4821'),
  ('CARD_PAYMENT', 'CUSTOMER', 1, 'TRANSACTION', 2, 'Seed card payment on account ****4821'),
  ('TRANSFER', 'CUSTOMER', 1, 'TRANSACTION', 3, 'Seed transfer from ****4821 to ****7312'),
  ('ACCOUNT_FROZEN', 'BANKER', NULL, 'ACCOUNT', 7, 'Account ****8890 frozen by banker simulation (demo data)');
