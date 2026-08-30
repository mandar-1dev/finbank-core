package com.corebank.entity;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Table(name = "cards")
@Getter
@Setter
@NoArgsConstructor
public class Card {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "account_id", nullable = false)
    private Account account;

    @Column(name = "card_number", nullable = false, unique = true, length = 19)
    private String cardNumber;

    @Column(name = "card_holder_name", nullable = false, length = 120)
    private String cardHolderName;

    @Column(name = "expiry_month", nullable = false)
    private Integer expiryMonth;

    @Column(name = "expiry_year", nullable = false)
    private Integer expiryYear;

    @Column(name = "cvv_hash", nullable = false)
    private String cvvHash;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private Status status = Status.ACTIVE;

    @Column(name = "spending_limit", nullable = false, precision = 19, scale = 4)
    private BigDecimal spendingLimit;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;

    @PrePersist
    void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
    }

    @PreUpdate
    void onUpdate() {
        updatedAt = LocalDateTime.now();
    }

    /** Masks all but the last 4 digits, e.g. **** **** **** 4821. */
    public String getMaskedCardNumber() {
        if (cardNumber == null || cardNumber.length() < 4) return "**** **** **** ****";
        return "**** **** **** " + cardNumber.substring(cardNumber.length() - 4);
    }

    public enum Status { ACTIVE, FROZEN, EXPIRED, CANCELLED }
}
