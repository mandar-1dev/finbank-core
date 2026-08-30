package com.corebank.service;

import com.corebank.dto.AccountResponse;
import com.corebank.entity.Account;
import com.corebank.exception.AccountNotFoundException;
import com.corebank.exception.UnauthorizedException;
import com.corebank.repository.AccountRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class AccountService {

    private final AccountRepository accountRepository;
    private final AuditService auditService;

    public AccountService(AccountRepository accountRepository, AuditService auditService) {
        this.accountRepository = accountRepository;
        this.auditService = auditService;
    }

    @Transactional(readOnly = true)
    public List<AccountResponse> getAccountsForCustomer(Long customerId) {
        return accountRepository.findByCustomerId(customerId).stream()
                .map(AccountResponse::from)
                .toList();
    }

    @Transactional(readOnly = true)
    public AccountResponse getAccountForCustomer(Long accountId, Long customerId) {
        Account account = getOwnedAccount(accountId, customerId);
        return AccountResponse.from(account);
    }

    /** Fetches an account and verifies it belongs to the requesting customer. */
    @Transactional(readOnly = true)
    public Account getOwnedAccount(Long accountId, Long customerId) {
        Account account = accountRepository.findById(accountId)
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        if (!account.getCustomer().getId().equals(customerId)) {
            throw new UnauthorizedException("This account does not belong to you");
        }
        return account;
    }

    @Transactional
    public void freezeAccount(Long accountId, Long adminId) {
        Account account = accountRepository.findById(accountId)
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        account.setStatus(Account.Status.FROZEN);
        accountRepository.save(account);
        auditService.log(adminId, "ACCOUNT_FREEZE", "ACCOUNT", accountId, "Account frozen by admin");
    }

    @Transactional
    public void unfreezeAccount(Long accountId, Long adminId) {
        Account account = accountRepository.findById(accountId)
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        account.setStatus(Account.Status.ACTIVE);
        accountRepository.save(account);
        auditService.log(adminId, "ACCOUNT_UNFREEZE", "ACCOUNT", accountId, "Account unfrozen by admin");
    }
}
