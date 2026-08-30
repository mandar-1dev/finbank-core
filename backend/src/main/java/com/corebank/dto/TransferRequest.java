package com.corebank.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record TransferRequest(
        @NotNull(message = "Source account id is required") Long sourceAccountId,
        String destinationAccountNumber,
        Long beneficiaryId,
        @NotNull(message = "Amount is required") @DecimalMin(value = "0.01", message = "Amount must be greater than zero") BigDecimal amount,
        String description,
        String idempotencyKey
) {}
