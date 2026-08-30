package com.corebank.repository;

import com.corebank.entity.Customer;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.Optional;

public interface CustomerRepository extends JpaRepository<Customer, Long> {
    Optional<Customer> findByEmail(String email);
    boolean existsByEmail(String email);
    Optional<Customer> findByCustomerNumber(String customerNumber);

    @Query("""
           SELECT c FROM Customer c
           WHERE (:query IS NULL OR LOWER(c.fullName) LIKE LOWER(CONCAT('%', :query, '%'))
                                  OR LOWER(c.email) LIKE LOWER(CONCAT('%', :query, '%'))
                                  OR c.customerNumber LIKE CONCAT('%', :query, '%'))
           ORDER BY c.createdAt DESC
           """)
    Page<Customer> search(@Param("query") String query, Pageable pageable);
}
