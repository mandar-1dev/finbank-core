package com.corebank.entity;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.math.BigDecimal;

@Entity
@Table(name = "account_types")
@Getter
@Setter
@NoArgsConstructor
public class AccountType {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, unique = true, length = 20)
    private Code code;

    @Column(length = 255)
    private String description;

    @Column(name = "min_balance", nullable = false, precision = 19, scale = 4)
    private BigDecimal minBalance;

    @Column(name = "interest_rate", nullable = false, precision = 5, scale = 2)
    private BigDecimal interestRate;

    public enum Code { SAVINGS, CURRENT }
}
