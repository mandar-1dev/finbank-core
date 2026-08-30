package com.corebank.dto;

import com.corebank.entity.Loan;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public record LoanResponse(
        Long id,
        BigDecimal principalAmount,
        BigDecimal interestRate,
        Integer tenureMonths,
        BigDecimal emiAmount,
        BigDecimal totalRepayment,
        BigDecimal outstandingAmount,
        String status,
        LocalDateTime appliedAt
) {
    public static LoanResponse from(Loan loan) {
        return new LoanResponse(loan.getId(), loan.getPrincipalAmount(), loan.getInterestRate(),
                loan.getTenureMonths(), loan.getEmiAmount(), loan.getTotalRepayment(),
                loan.getOutstandingAmount(), loan.getStatus().name(), loan.getAppliedAt());
    }
}
