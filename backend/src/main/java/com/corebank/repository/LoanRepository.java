package com.corebank.repository;

import com.corebank.entity.Loan;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface LoanRepository extends JpaRepository<Loan, Long> {
    List<Loan> findByCustomerId(Long customerId);
    long countByStatus(Loan.Status status);
}
