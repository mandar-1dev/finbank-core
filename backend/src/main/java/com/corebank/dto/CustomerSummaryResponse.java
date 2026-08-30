package com.corebank.dto;

import com.corebank.entity.Customer;

public record CustomerSummaryResponse(
        Long id,
        String customerNumber,
        String fullName,
        String email,
        String phone,
        String role,
        String status
) {
    public static CustomerSummaryResponse from(Customer c) {
        return new CustomerSummaryResponse(c.getId(), c.getCustomerNumber(), c.getFullName(),
                c.getEmail(), c.getPhone(), c.getRole().name(), c.getStatus().name());
    }
}
