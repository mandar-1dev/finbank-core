package com.corebank.dto;

import com.corebank.entity.Account;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public record AccountResponse(
        Long id,
        String maskedAccountNumber,
        String accountType,
        BigDecimal balance,
        BigDecimal availableBalance,
        String status,
        LocalDateTime openedAt
) {
    public static AccountResponse from(Account account) {
        return new AccountResponse(
                account.getId(),
                account.getMaskedAccountNumber(),
                account.getAccountType().getCode().name(),
                account.getBalance(),
                account.getAvailableBalance(),
                account.getStatus().name(),
                account.getOpenedAt()
        );
    }
}
