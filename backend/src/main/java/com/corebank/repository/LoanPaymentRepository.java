package com.corebank.repository;

import com.corebank.entity.LoanPayment;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface LoanPaymentRepository extends JpaRepository<LoanPayment, Long> {
    List<LoanPayment> findByLoanIdOrderByPaymentDateDesc(Long loanId);
}
