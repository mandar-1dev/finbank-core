package com.corebank.service;

import com.corebank.dto.*;
import com.corebank.entity.*;
import com.corebank.exception.*;
import com.corebank.repository.AccountRepository;
import com.corebank.repository.BeneficiaryRepository;
import com.corebank.repository.LedgerEntryRepository;
import com.corebank.repository.TransactionRepository;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.UUID;
import java.util.concurrent.ThreadLocalRandom;

/**
 * Owns every operation that moves money between accounts: deposits,
 * withdrawals and transfers. This is the most safety-critical service in
 * the platform, so every method here:
 *
 *   1. Runs inside a single @Transactional boundary (all-or-nothing).
 *   2. Locks the account row(s) involved with SELECT ... FOR UPDATE via
 *      {@link AccountRepository#findByIdForUpdate}, always acquiring
 *      locks in ascending account-id order to avoid deadlocks between
 *      two transfers that touch the same pair of accounts in reverse.
 *   3. Writes a balanced pair of {@link LedgerEntry} rows (one DEBIT,
 *      one CREDIT) for every money movement.
 *   4. Is idempotent: a transfer submitted twice with the same
 *      Idempotency-Key returns the original result instead of moving
 *      money twice.
 */
@Service
public class TransactionService {

    private static final DateTimeFormatter REF_FORMAT = DateTimeFormatter.ofPattern("yyyyMMdd");

    private final AccountRepository accountRepository;
    private final TransactionRepository transactionRepository;
    private final LedgerEntryRepository ledgerEntryRepository;
    private final BeneficiaryRepository beneficiaryRepository;
    private final AuditService auditService;

    public TransactionService(AccountRepository accountRepository,
                               TransactionRepository transactionRepository,
                               LedgerEntryRepository ledgerEntryRepository,
                               BeneficiaryRepository beneficiaryRepository,
                               AuditService auditService) {
        this.accountRepository = accountRepository;
        this.transactionRepository = transactionRepository;
        this.ledgerEntryRepository = ledgerEntryRepository;
        this.beneficiaryRepository = beneficiaryRepository;
        this.auditService = auditService;
    }

    // ------------------------------------------------------------------
    // DEPOSIT
    // ------------------------------------------------------------------
    @Transactional
    public TransactionResponse deposit(DepositRequest request, Long customerId) {
        Account account = accountRepository.findByIdForUpdate(request.accountId())
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        assertOwnership(account, customerId);
        assertNotClosed(account);
        // Deposits are allowed even on a frozen account per the freeze policy (withdrawals/transfers/card
        // payments are blocked; incoming credits are not) — see AccountFreezingSpec in docs/architecture.md.

        BigDecimal newBalance = account.getBalance().add(request.amount());
        account.setBalance(newBalance);
        account.setAvailableBalance(account.getAvailableBalance().add(request.amount()));
        accountRepository.save(account);

        Transaction txn = newTransaction(Transaction.Type.DEPOSIT, null, account, request.amount(),
                request.description() == null ? "Cash Deposit Simulation" : request.description(), null);
        txn.setStatus(Transaction.Status.SUCCESS);
        transactionRepository.save(txn);

        writeLedgerEntry(txn, account, LedgerEntry.EntryType.CREDIT, request.amount(), newBalance);

        auditService.log(customerId, "DEPOSIT", "TRANSACTION", txn.getId(),
                "Deposited " + request.amount() + " into account " + account.getMaskedAccountNumber());

        return TransactionResponse.from(txn);
    }

    // ------------------------------------------------------------------
    // WITHDRAWAL
    // ------------------------------------------------------------------
    @Transactional
    public TransactionResponse withdraw(WithdrawRequest request, Long customerId) {
        Account account = accountRepository.findByIdForUpdate(request.accountId())
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        assertOwnership(account, customerId);
        assertActive(account);

        if (account.getAvailableBalance().compareTo(request.amount()) < 0) {
            throw new InsufficientBalanceException("Insufficient account balance");
        }

        BigDecimal newBalance = account.getBalance().subtract(request.amount());
        account.setBalance(newBalance);
        account.setAvailableBalance(account.getAvailableBalance().subtract(request.amount()));
        accountRepository.save(account);

        Transaction txn = newTransaction(Transaction.Type.WITHDRAWAL, account, null, request.amount(),
                request.description() == null ? "Withdrawal Simulation" : request.description(), null);
        txn.setStatus(Transaction.Status.SUCCESS);
        transactionRepository.save(txn);

        writeLedgerEntry(txn, account, LedgerEntry.EntryType.DEBIT, request.amount(), newBalance);

        auditService.log(customerId, "WITHDRAWAL", "TRANSACTION", txn.getId(),
                "Withdrew " + request.amount() + " from account " + account.getMaskedAccountNumber());

        return TransactionResponse.from(txn);
    }

    // ------------------------------------------------------------------
    // TRANSFER — the most important operation in the platform
    // ------------------------------------------------------------------
    @Transactional
    public TransactionResponse transfer(TransferRequest request, Long customerId) {

        // --- Idempotency check -----------------------------------------
        // If this exact request (by client-supplied key) already succeeded,
        // return the original result instead of moving money again.
        if (request.idempotencyKey() != null && !request.idempotencyKey().isBlank()) {
            var existing = transactionRepository.findByIdempotencyKey(request.idempotencyKey());
            if (existing.isPresent()) {
                return TransactionResponse.from(existing.get());
            }
        }

        if (request.amount().compareTo(BigDecimal.ZERO) <= 0) {
            throw new InvalidTransactionException("Transfer amount must be greater than zero");
        }

        String destinationAccountNumber = resolveDestinationAccountNumber(request, customerId);

        Account destinationLookup = accountRepository.findByAccountNumber(destinationAccountNumber)
                .orElseThrow(() -> new AccountNotFoundException("Destination account not found"));

        if (destinationLookup.getId().equals(request.sourceAccountId())) {
            throw new InvalidTransactionException("Cannot transfer to the same account");
        }

        // --- Lock both accounts in a fixed order (ascending id) to avoid
        // deadlocks when two transfers cross the same pair of accounts.
        Long firstLockId = Math.min(request.sourceAccountId(), destinationLookup.getId());
        Long secondLockId = Math.max(request.sourceAccountId(), destinationLookup.getId());

        Account first = accountRepository.findByIdForUpdate(firstLockId)
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        Account second = accountRepository.findByIdForUpdate(secondLockId)
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));

        Account source = first.getId().equals(request.sourceAccountId()) ? first : second;
        Account destination = first.getId().equals(request.sourceAccountId()) ? second : first;

        assertOwnership(source, customerId);
        assertActive(source);
        assertNotClosed(destination);
        if (destination.getStatus() == Account.Status.FROZEN) {
            throw new AccountFrozenException("Destination account is frozen and cannot receive transfers");
        }

        if (source.getAvailableBalance().compareTo(request.amount()) < 0) {
            throw new InsufficientBalanceException("Insufficient account balance");
        }

        BigDecimal sourceNewBalance = source.getBalance().subtract(request.amount());
        BigDecimal destinationNewBalance = destination.getBalance().add(request.amount());

        source.setBalance(sourceNewBalance);
        source.setAvailableBalance(source.getAvailableBalance().subtract(request.amount()));
        destination.setBalance(destinationNewBalance);
        destination.setAvailableBalance(destination.getAvailableBalance().add(request.amount()));

        accountRepository.save(source);
        accountRepository.save(destination);

        Transaction txn = newTransaction(Transaction.Type.TRANSFER, source, destination, request.amount(),
                request.description() == null ? "Fund Transfer" : request.description(),
                request.idempotencyKey());
        txn.setStatus(Transaction.Status.SUCCESS);

        try {
            transactionRepository.save(txn);
        } catch (org.springframework.dao.DataIntegrityViolationException ex) {
            // A concurrent request with the same idempotency key won the race —
            // return that transaction's result rather than a duplicate transfer.
            if (request.idempotencyKey() != null) {
                return transactionRepository.findByIdempotencyKey(request.idempotencyKey())
                        .map(TransactionResponse::from)
                        .orElseThrow(() -> new DuplicateTransactionException("Duplicate transfer request"));
            }
            throw new DuplicateTransactionException("Duplicate transfer request");
        }

        writeLedgerEntry(txn, source, LedgerEntry.EntryType.DEBIT, request.amount(), sourceNewBalance);
        writeLedgerEntry(txn, destination, LedgerEntry.EntryType.CREDIT, request.amount(), destinationNewBalance);

        auditService.log(customerId, "TRANSFER", "TRANSACTION", txn.getId(),
                "Transferred " + request.amount() + " from " + source.getMaskedAccountNumber() +
                        " to " + destination.getMaskedAccountNumber());

        return TransactionResponse.from(txn);
    }

    // ------------------------------------------------------------------
    // HISTORY
    // ------------------------------------------------------------------
    @Transactional(readOnly = true)
    public Page<TransactionResponse> getHistory(Long accountId, Long customerId, Transaction.Type type,
                                                 LocalDateTime from, LocalDateTime to, Pageable pageable) {
        Account account = accountRepository.findById(accountId)
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        assertOwnership(account, customerId);
        return transactionRepository.findHistoryForAccount(accountId, type, from, to, pageable)
                .map(TransactionResponse::from);
    }

    // ------------------------------------------------------------------
    // Internal helpers
    // ------------------------------------------------------------------

    private Transaction newTransaction(Transaction.Type type, Account source, Account destination,
                                        BigDecimal amount, String description, String idempotencyKey) {
        Transaction txn = new Transaction();
        txn.setTransactionRef(generateTransactionRef());
        txn.setIdempotencyKey(idempotencyKey);
        txn.setType(type);
        txn.setSourceAccount(source);
        txn.setDestinationAccount(destination);
        txn.setAmount(amount);
        txn.setDescription(description);
        return txn;
    }

    private void writeLedgerEntry(Transaction txn, Account account, LedgerEntry.EntryType type,
                                   BigDecimal amount, BigDecimal balanceAfter) {
        LedgerEntry entry = new LedgerEntry();
        entry.setTransaction(txn);
        entry.setAccount(account);
        entry.setEntryType(type);
        entry.setAmount(amount);
        entry.setBalanceAfter(balanceAfter);
        ledgerEntryRepository.save(entry);
    }

    private String generateTransactionRef() {
        String datePart = LocalDateTime.now().format(REF_FORMAT);
        String randomPart = String.valueOf(ThreadLocalRandom.current().nextInt(100000, 999999));
        return "TXN" + datePart + randomPart;
    }

    private void assertOwnership(Account account, Long customerId) {
        if (!account.getCustomer().getId().equals(customerId)) {
            throw new UnauthorizedException("This account does not belong to you");
        }
    }

    private void assertActive(Account account) {
        if (account.getStatus() == Account.Status.FROZEN) {
            throw new AccountFrozenException("Account is frozen");
        }
        assertNotClosed(account);
    }

    private void assertNotClosed(Account account) {
        if (account.getStatus() == Account.Status.CLOSED) {
            throw new InvalidTransactionException("Account is closed");
        }
    }

    /**
     * Resolves the destination account number either directly (when the
     * caller passed one) or via a saved beneficiary owned by the caller —
     * the frontend never needs to know or send an unmasked account number
     * for a beneficiary it only has the masked display value for.
     */
    private String resolveDestinationAccountNumber(TransferRequest request, Long customerId) {
        if (request.beneficiaryId() != null) {
            var beneficiary = beneficiaryRepository.findByIdAndCustomerId(request.beneficiaryId(), customerId)
                    .orElseThrow(() -> new com.corebank.exception.BeneficiaryNotFoundException("Beneficiary not found"));
            return beneficiary.getAccountNumber();
        }
        if (request.destinationAccountNumber() != null && !request.destinationAccountNumber().isBlank()) {
            return request.destinationAccountNumber();
        }
        throw new InvalidTransactionException("Either a beneficiaryId or a destination account number is required");
    }
}
