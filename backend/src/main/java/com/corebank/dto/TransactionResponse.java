package com.corebank.dto;

import com.corebank.entity.Transaction;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public record TransactionResponse(
        Long id,
        String transactionRef,
        String type,
        String sourceAccountMasked,
        String destinationAccountMasked,
        BigDecimal amount,
        BigDecimal fee,
        String description,
        String status,
        String failureReason,
        LocalDateTime createdAt
) {
    public static TransactionResponse from(Transaction txn) {
        return new TransactionResponse(
                txn.getId(),
                txn.getTransactionRef(),
                txn.getType().name(),
                txn.getSourceAccount() != null ? txn.getSourceAccount().getMaskedAccountNumber() : null,
                txn.getDestinationAccount() != null ? txn.getDestinationAccount().getMaskedAccountNumber() : null,
                txn.getAmount(),
                txn.getFee(),
                txn.getDescription(),
                txn.getStatus().name(),
                txn.getFailureReason(),
                txn.getCreatedAt()
        );
    }
}
