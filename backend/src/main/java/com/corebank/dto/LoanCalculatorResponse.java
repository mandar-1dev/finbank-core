package com.corebank.dto;

import java.math.BigDecimal;

public record LoanCalculatorResponse(
        BigDecimal emi,
        BigDecimal totalInterest,
        BigDecimal totalRepayment
) {}
