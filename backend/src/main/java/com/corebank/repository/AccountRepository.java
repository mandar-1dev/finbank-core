package com.corebank.repository;

import com.corebank.entity.Account;
import jakarta.persistence.LockModeType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Lock;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.List;
import java.util.Optional;

public interface AccountRepository extends JpaRepository<Account, Long> {

    Optional<Account> findByAccountNumber(String accountNumber);

    List<Account> findByCustomerId(Long customerId);

    /**
     * Locks the account row (SELECT ... FOR UPDATE) so that two concurrent
     * transfers touching the same account cannot both read a stale balance.
     * MUST be called from within an active @Transactional method, and
     * accounts involved in a transfer must always be locked in a
     * consistent order (e.g. by ascending id) to avoid deadlocks.
     */
    @Lock(LockModeType.PESSIMISTIC_WRITE)
    @Query("SELECT a FROM Account a WHERE a.id = :id")
    Optional<Account> findByIdForUpdate(@Param("id") Long id);
}
