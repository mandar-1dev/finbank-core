package com.corebank.dto;

public record CreditScoreResponse(
        int score,
        String rating,
        boolean simulated
) {
    public static CreditScoreResponse of(int score, String rating) {
        return new CreditScoreResponse(score, rating, true);
    }
}
