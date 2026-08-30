package com.corebank.service;

import com.corebank.dto.BeneficiaryRequest;
import com.corebank.dto.BeneficiaryResponse;
import com.corebank.entity.Beneficiary;
import com.corebank.exception.AccountNotFoundException;
import com.corebank.exception.BeneficiaryNotFoundException;
import com.corebank.exception.InvalidTransactionException;
import com.corebank.repository.AccountRepository;
import com.corebank.repository.BeneficiaryRepository;
import com.corebank.repository.CustomerRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class BeneficiaryService {

    private final BeneficiaryRepository beneficiaryRepository;
    private final AccountRepository accountRepository;
    private final CustomerRepository customerRepository;
    private final AuditService auditService;

    public BeneficiaryService(BeneficiaryRepository beneficiaryRepository, AccountRepository accountRepository,
                               CustomerRepository customerRepository, AuditService auditService) {
        this.beneficiaryRepository = beneficiaryRepository;
        this.accountRepository = accountRepository;
        this.customerRepository = customerRepository;
        this.auditService = auditService;
    }

    @Transactional(readOnly = true)
    public List<BeneficiaryResponse> list(Long customerId) {
        return beneficiaryRepository.findByCustomerId(customerId).stream()
                .map(BeneficiaryResponse::from)
                .toList();
    }

    @Transactional
    public BeneficiaryResponse add(BeneficiaryRequest request, Long customerId) {
        if (!accountRepository.findByAccountNumber(request.accountNumber()).isPresent()) {
            throw new AccountNotFoundException("No CoreBank account exists with that account number");
        }

        Beneficiary beneficiary = new Beneficiary();
        // getReferenceById returns a lazy proxy (no extra SELECT) that
        // Hibernate resolves to a foreign key on save.
        beneficiary.setCustomer(customerRepository.getReferenceById(customerId));
        beneficiary.setName(request.name());
        beneficiary.setAccountNumber(request.accountNumber());
        beneficiary.setIfscCode(request.ifscCode());
        beneficiary.setNickname(request.nickname());

        try {
            beneficiary = beneficiaryRepository.save(beneficiary);
        } catch (org.springframework.dao.DataIntegrityViolationException ex) {
            throw new InvalidTransactionException("This beneficiary has already been added");
        }

        auditService.log(customerId, "BENEFICIARY_ADD", "BENEFICIARY", beneficiary.getId(),
                "Added beneficiary " + request.name());

        return BeneficiaryResponse.from(beneficiary);
    }

    @Transactional
    public void remove(Long beneficiaryId, Long customerId) {
        Beneficiary beneficiary = beneficiaryRepository.findByIdAndCustomerId(beneficiaryId, customerId)
                .orElseThrow(() -> new BeneficiaryNotFoundException("Beneficiary not found"));
        beneficiaryRepository.delete(beneficiary);
        auditService.log(customerId, "BENEFICIARY_DELETE", "BENEFICIARY", beneficiaryId,
                "Removed beneficiary " + beneficiary.getName());
    }
}
