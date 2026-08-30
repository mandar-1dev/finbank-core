package com.corebank.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record CardPaymentRequest(
        @NotNull(message = "Card id is required") Long cardId,
        @NotBlank(message = "Merchant name is required") String merchant,
        @NotNull(message = "Amount is required") @DecimalMin(value = "0.01", message = "Amount must be greater than zero") BigDecimal amount
) {}
