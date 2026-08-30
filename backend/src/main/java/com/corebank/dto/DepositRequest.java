package com.corebank.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record DepositRequest(
        @NotNull(message = "Account id is required") Long accountId,
        @NotNull(message = "Amount is required") @DecimalMin(value = "0.01", message = "Amount must be greater than zero") BigDecimal amount,
        String description
) {}
