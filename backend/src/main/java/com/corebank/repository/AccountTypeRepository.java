package com.corebank.repository;

import com.corebank.entity.AccountType;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;

public interface AccountTypeRepository extends JpaRepository<AccountType, Long> {
    Optional<AccountType> findByCode(AccountType.Code code);
}
