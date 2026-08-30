package com.corebank.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;

import java.math.BigDecimal;

public record LoanApplicationRequest(
        @NotNull Long accountId,
        @NotNull @DecimalMin(value = "1000", message = "Loan amount must be at least 1000") BigDecimal principal,
        @NotNull @DecimalMin(value = "0.1") BigDecimal annualInterestRate,
        @NotNull @Positive Integer tenureMonths
) {}
