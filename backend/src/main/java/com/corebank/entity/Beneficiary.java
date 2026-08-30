package com.corebank.entity;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.time.LocalDateTime;

@Entity
@Table(name = "beneficiaries")
@Getter
@Setter
@NoArgsConstructor
public class Beneficiary {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "customer_id", nullable = false)
    private Customer customer;

    @Column(nullable = false, length = 120)
    private String name;

    @Column(name = "account_number", nullable = false, length = 20)
    private String accountNumber;

    @Column(name = "ifsc_code", nullable = false, length = 15)
    private String ifscCode;

    @Column(length = 60)
    private String nickname;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @PrePersist
    void onCreate() {
        createdAt = LocalDateTime.now();
    }
}
