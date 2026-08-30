package com.corebank.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;

import java.math.BigDecimal;

public record LoanCalculatorRequest(
        @NotNull @DecimalMin(value = "1000", message = "Loan amount must be at least 1000") BigDecimal principal,
        @NotNull @DecimalMin(value = "0.1", message = "Interest rate must be greater than zero") BigDecimal annualInterestRate,
        @NotNull @Positive(message = "Tenure must be a positive number of months") Integer tenureMonths
) {}
