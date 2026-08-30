package com.corebank.dto;

import jakarta.validation.constraints.NotBlank;

public record BeneficiaryRequest(
        @NotBlank(message = "Name is required") String name,
        @NotBlank(message = "Account number is required") String accountNumber,
        @NotBlank(message = "IFSC code is required") String ifscCode,
        String nickname
) {}
