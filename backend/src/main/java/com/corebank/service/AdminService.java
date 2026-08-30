package com.corebank.service;

import com.corebank.dto.AccountResponse;
import com.corebank.dto.AdminDashboardResponse;
import com.corebank.dto.AuditLogResponse;
import com.corebank.dto.CustomerSummaryResponse;
import com.corebank.entity.Account;
import com.corebank.entity.Loan;
import com.corebank.entity.Transaction;
import com.corebank.exception.CustomerNotFoundException;
import com.corebank.repository.*;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

@Service
public class AdminService {

    private final CustomerRepository customerRepository;
    private final AccountRepository accountRepository;
    private final TransactionRepository transactionRepository;
    private final LoanRepository loanRepository;
    private final AuditLogRepository auditLogRepository;

    public AdminService(CustomerRepository customerRepository, AccountRepository accountRepository,
                         TransactionRepository transactionRepository, LoanRepository loanRepository,
                         AuditLogRepository auditLogRepository) {
        this.customerRepository = customerRepository;
        this.accountRepository = accountRepository;
        this.transactionRepository = transactionRepository;
        this.loanRepository = loanRepository;
        this.auditLogRepository = auditLogRepository;
    }

    @Transactional(readOnly = true)
    public AdminDashboardResponse dashboard() {
        List<Account> allAccounts = accountRepository.findAll();
        long active = allAccounts.stream().filter(a -> a.getStatus() == Account.Status.ACTIVE).count();
        long frozen = allAccounts.stream().filter(a -> a.getStatus() == Account.Status.FROZEN).count();

        LocalDateTime startOfDay = LocalDate.now().atStartOfDay();
        long txnToday = transactionRepository.countByCreatedAtAfter(startOfDay);

        BigDecimal volumeToday = transactionRepository.findAll().stream()
                .filter(t -> t.getCreatedAt().isAfter(startOfDay) && t.getStatus() == Transaction.Status.SUCCESS)
                .map(Transaction::getAmount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        long pendingLoans = loanRepository.countByStatus(Loan.Status.PENDING);

        // "Suspicious" here is a simple simulated heuristic: any single
        // transaction above 200,000 in the last 24 hours.
        long suspicious = transactionRepository.findAll().stream()
                .filter(t -> t.getCreatedAt().isAfter(LocalDateTime.now().minusHours(24)))
                .filter(t -> t.getAmount().compareTo(BigDecimal.valueOf(200000)) > 0)
                .count();

        return new AdminDashboardResponse(
                customerRepository.count(),
                allAccounts.size(),
                active,
                frozen,
                txnToday,
                volumeToday,
                pendingLoans,
                suspicious
        );
    }

    @Transactional(readOnly = true)
    public Page<AuditLogResponse> auditLogs(Pageable pageable) {
        return auditLogRepository.findAllByOrderByTimestampDesc(pageable).map(AuditLogResponse::from);
    }

    @Transactional(readOnly = true)
    public Page<CustomerSummaryResponse> searchCustomers(String query, Pageable pageable) {
        return customerRepository.search(query, pageable).map(CustomerSummaryResponse::from);
    }

    @Transactional(readOnly = true)
    public List<AccountResponse> customerAccounts(Long customerId) {
        if (!customerRepository.existsById(customerId)) {
            throw new CustomerNotFoundException("Customer not found");
        }
        return accountRepository.findByCustomerId(customerId).stream().map(AccountResponse::from).toList();
    }
}
