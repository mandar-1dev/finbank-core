package com.corebank.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;

public record SpendingLimitRequest(
        @NotNull @DecimalMin(value = "0.01", message = "Limit must be greater than zero") BigDecimal spendingLimit
) {}
