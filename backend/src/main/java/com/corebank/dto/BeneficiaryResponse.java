package com.corebank.dto;

import com.corebank.entity.Beneficiary;

public record BeneficiaryResponse(
        Long id,
        String name,
        String maskedAccountNumber,
        String ifscCode,
        String nickname
) {
    public static BeneficiaryResponse from(Beneficiary b) {
        String acc = b.getAccountNumber();
        String masked = acc.length() >= 4 ? "****" + acc.substring(acc.length() - 4) : "****";
        return new BeneficiaryResponse(b.getId(), b.getName(), masked, b.getIfscCode(), b.getNickname());
    }
}
