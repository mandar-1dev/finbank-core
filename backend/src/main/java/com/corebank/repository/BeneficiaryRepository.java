package com.corebank.repository;

import com.corebank.entity.Beneficiary;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.Optional;

public interface BeneficiaryRepository extends JpaRepository<Beneficiary, Long> {
    List<Beneficiary> findByCustomerId(Long customerId);
    Optional<Beneficiary> findByIdAndCustomerId(Long id, Long customerId);
}
