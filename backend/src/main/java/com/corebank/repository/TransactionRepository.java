package com.corebank.repository;

import com.corebank.entity.Transaction;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.LocalDateTime;
import java.util.Optional;

public interface TransactionRepository extends JpaRepository<Transaction, Long> {

    Optional<Transaction> findByIdempotencyKey(String idempotencyKey);

    Optional<Transaction> findByTransactionRef(String transactionRef);

    @Query("""
           SELECT t FROM Transaction t
           WHERE (t.sourceAccount.id = :accountId OR t.destinationAccount.id = :accountId)
           AND (:type IS NULL OR t.type = :type)
           AND (:from IS NULL OR t.createdAt >= :from)
           AND (:to IS NULL OR t.createdAt <= :to)
           ORDER BY t.createdAt DESC
           """)
    Page<Transaction> findHistoryForAccount(@Param("accountId") Long accountId,
                                             @Param("type") Transaction.Type type,
                                             @Param("from") LocalDateTime from,
                                             @Param("to") LocalDateTime to,
                                             Pageable pageable);

    long countByCreatedAtAfter(LocalDateTime after);
}
