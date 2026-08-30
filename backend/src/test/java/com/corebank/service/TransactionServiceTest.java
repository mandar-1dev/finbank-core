package com.corebank.service;

import com.corebank.dto.DepositRequest;
import com.corebank.dto.TransferRequest;
import com.corebank.dto.WithdrawRequest;
import com.corebank.entity.Account;
import com.corebank.entity.AccountType;
import com.corebank.entity.Customer;
import com.corebank.entity.LedgerEntry;
import com.corebank.entity.Transaction;
import com.corebank.exception.AccountFrozenException;
import com.corebank.exception.InsufficientBalanceException;
import com.corebank.exception.UnauthorizedException;
import com.corebank.repository.AccountRepository;
import com.corebank.repository.LedgerEntryRepository;
import com.corebank.repository.TransactionRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

/**
 * Unit tests for the money-movement logic in TransactionService.
 * Repositories are mocked so these tests run without a real database and
 * focus purely on the business rules: balance checks, frozen-account
 * checks, ownership checks and idempotency.
 */
@ExtendWith(MockitoExtension.class)
class TransactionServiceTest {

    @Mock private AccountRepository accountRepository;
    @Mock private TransactionRepository transactionRepository;
    @Mock private LedgerEntryRepository ledgerEntryRepository;
    @Mock private com.corebank.repository.BeneficiaryRepository beneficiaryRepository;
    @Mock private AuditService auditService;

    private TransactionService transactionService;

    private Customer owner;
    private Account sourceAccount;
    private Account destinationAccount;

    @BeforeEach
    void setUp() {
        transactionService = new TransactionService(accountRepository, transactionRepository,
                ledgerEntryRepository, beneficiaryRepository, auditService);

        owner = new Customer();
        owner.setId(1L);

        Customer otherCustomer = new Customer();
        otherCustomer.setId(2L);

        AccountType savings = new AccountType();
        savings.setCode(AccountType.Code.SAVINGS);

        sourceAccount = new Account();
        sourceAccount.setId(10L);
        sourceAccount.setAccountNumber("4000100014821");
        sourceAccount.setCustomer(owner);
        sourceAccount.setAccountType(savings);
        sourceAccount.setBalance(new BigDecimal("10000.00"));
        sourceAccount.setAvailableBalance(new BigDecimal("10000.00"));
        sourceAccount.setStatus(Account.Status.ACTIVE);

        destinationAccount = new Account();
        destinationAccount.setId(20L);
        destinationAccount.setAccountNumber("4000100027312");
        destinationAccount.setCustomer(otherCustomer);
        destinationAccount.setAccountType(savings);
        destinationAccount.setBalance(new BigDecimal("5000.00"));
        destinationAccount.setAvailableBalance(new BigDecimal("5000.00"));
        destinationAccount.setStatus(Account.Status.ACTIVE);
    }

    @Test
    void deposit_increasesBalanceAndCreatesLedgerEntry() {
        when(accountRepository.findByIdForUpdate(10L)).thenReturn(Optional.of(sourceAccount));
        when(transactionRepository.save(any(Transaction.class))).thenAnswer(inv -> inv.getArgument(0));

        var response = transactionService.deposit(new DepositRequest(10L, new BigDecimal("500.00"), "Test deposit"), 1L);

        assertThat(sourceAccount.getBalance()).isEqualByComparingTo("10500.00");
        assertThat(response.status()).isEqualTo("SUCCESS");
        verify(ledgerEntryRepository).save(argThat(entry -> entry.getEntryType() == LedgerEntry.EntryType.CREDIT));
    }

    @Test
    void withdraw_failsWithInsufficientBalance() {
        when(accountRepository.findByIdForUpdate(10L)).thenReturn(Optional.of(sourceAccount));

        assertThatThrownBy(() ->
                transactionService.withdraw(new WithdrawRequest(10L, new BigDecimal("50000.00"), null), 1L)
        ).isInstanceOf(InsufficientBalanceException.class);

        // balance must be unchanged on failure
        assertThat(sourceAccount.getBalance()).isEqualByComparingTo("10000.00");
    }

    @Test
    void withdraw_failsOnFrozenAccount() {
        sourceAccount.setStatus(Account.Status.FROZEN);
        when(accountRepository.findByIdForUpdate(10L)).thenReturn(Optional.of(sourceAccount));

        assertThatThrownBy(() ->
                transactionService.withdraw(new WithdrawRequest(10L, new BigDecimal("100.00"), null), 1L)
        ).isInstanceOf(AccountFrozenException.class);
    }

    @Test
    void withdraw_failsWhenAccountNotOwnedByCaller() {
        when(accountRepository.findByIdForUpdate(10L)).thenReturn(Optional.of(sourceAccount));

        assertThatThrownBy(() ->
                transactionService.withdraw(new WithdrawRequest(10L, new BigDecimal("100.00"), null), 999L)
        ).isInstanceOf(UnauthorizedException.class);
    }

    @Test
    void transfer_movesMoneyBetweenAccountsAtomically() {
        when(accountRepository.findByAccountNumber("4000100027312")).thenReturn(Optional.of(destinationAccount));
        when(accountRepository.findByIdForUpdate(10L)).thenReturn(Optional.of(sourceAccount));
        when(accountRepository.findByIdForUpdate(20L)).thenReturn(Optional.of(destinationAccount));
        when(transactionRepository.save(any(Transaction.class))).thenAnswer(inv -> inv.getArgument(0));

        var response = transactionService.transfer(
                new TransferRequest(10L, "4000100027312", null, new BigDecimal("2000.00"), "Rent", null), 1L);

        assertThat(sourceAccount.getBalance()).isEqualByComparingTo("8000.00");
        assertThat(destinationAccount.getBalance()).isEqualByComparingTo("7000.00");
        assertThat(response.status()).isEqualTo("SUCCESS");

        // one DEBIT entry for the source, one CREDIT entry for the destination
        verify(ledgerEntryRepository, times(2)).save(any(LedgerEntry.class));
    }

    @Test
    void transfer_isIdempotent_returnsExistingResultOnDuplicateKey() {
        Transaction existing = new Transaction();
        existing.setTransactionRef("TXN20260101000001");
        existing.setType(Transaction.Type.TRANSFER);
        existing.setAmount(new BigDecimal("2000.00"));
        existing.setStatus(Transaction.Status.SUCCESS);

        when(transactionRepository.findByIdempotencyKey("dup-key")).thenReturn(Optional.of(existing));

        var response = transactionService.transfer(
                new TransferRequest(10L, "4000100027312", null, new BigDecimal("2000.00"), "Rent", "dup-key"), 1L);

        assertThat(response.transactionRef()).isEqualTo("TXN20260101000001");
        // no account lookups should happen once the idempotency key short-circuits the request
        verify(accountRepository, never()).findByIdForUpdate(any());
    }

    @Test
    void transfer_failsWithInsufficientBalance() {
        when(accountRepository.findByAccountNumber("4000100027312")).thenReturn(Optional.of(destinationAccount));
        when(accountRepository.findByIdForUpdate(10L)).thenReturn(Optional.of(sourceAccount));
        when(accountRepository.findByIdForUpdate(20L)).thenReturn(Optional.of(destinationAccount));

        assertThatThrownBy(() ->
                transactionService.transfer(
                        new TransferRequest(10L, "4000100027312", null, new BigDecimal("999999.00"), null, null), 1L)
        ).isInstanceOf(InsufficientBalanceException.class);

        // Neither balance should move when the transfer is rejected.
        assertThat(sourceAccount.getBalance()).isEqualByComparingTo("10000.00");
        assertThat(destinationAccount.getBalance()).isEqualByComparingTo("5000.00");
    }
}
