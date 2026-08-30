package com.corebank.repository;

import com.corebank.entity.LedgerEntry;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface LedgerEntryRepository extends JpaRepository<LedgerEntry, Long> {
    List<LedgerEntry> findByAccountIdOrderByCreatedAtDesc(Long accountId);
    List<LedgerEntry> findByTransactionId(Long transactionId);
}
