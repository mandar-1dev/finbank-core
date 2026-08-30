package com.corebank.dto;

public record LoginResponse(
        String token,
        String tokenType,
        Long customerId,
        String fullName,
        String role
) {
    public static LoginResponse of(String token, Long customerId, String fullName, String role) {
        return new LoginResponse(token, "Bearer", customerId, fullName, role);
    }
}
