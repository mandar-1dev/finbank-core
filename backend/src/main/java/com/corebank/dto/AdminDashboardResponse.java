package com.corebank.dto;

public record AdminDashboardResponse(
        long totalCustomers,
        long totalAccounts,
        long activeAccounts,
        long frozenAccounts,
        long transactionsToday,
        java.math.BigDecimal transactionVolumeToday,
        long pendingLoans,
        long suspiciousTransactions
) {}
