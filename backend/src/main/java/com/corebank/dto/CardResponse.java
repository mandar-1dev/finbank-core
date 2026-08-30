package com.corebank.dto;

import com.corebank.entity.Card;

import java.math.BigDecimal;

public record CardResponse(
        Long id,
        String maskedCardNumber,
        String cardHolderName,
        String expiry,
        String status,
        BigDecimal spendingLimit
) {
    public static CardResponse from(Card card) {
        String expiry = String.format("%02d/%d", card.getExpiryMonth(), card.getExpiryYear() % 100);
        return new CardResponse(card.getId(), card.getMaskedCardNumber(), card.getCardHolderName(),
                expiry, card.getStatus().name(), card.getSpendingLimit());
    }
}
