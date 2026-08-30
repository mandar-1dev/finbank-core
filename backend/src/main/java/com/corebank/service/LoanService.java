package com.corebank.service;

import com.corebank.dto.*;
import com.corebank.entity.*;
import com.corebank.exception.AccountNotFoundException;
import com.corebank.repository.AccountRepository;
import com.corebank.repository.LedgerEntryRepository;
import com.corebank.repository.LoanRepository;
import com.corebank.repository.TransactionRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.MathContext;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.concurrent.ThreadLocalRandom;

@Service
public class LoanService {

    private static final DateTimeFormatter REF_FORMAT = DateTimeFormatter.ofPattern("yyyyMMdd");

    private final LoanRepository loanRepository;
    private final AccountRepository accountRepository;
    private final TransactionRepository transactionRepository;
    private final LedgerEntryRepository ledgerEntryRepository;
    private final AuditService auditService;

    public LoanService(LoanRepository loanRepository, AccountRepository accountRepository,
                        TransactionRepository transactionRepository, LedgerEntryRepository ledgerEntryRepository,
                        AuditService auditService) {
        this.loanRepository = loanRepository;
        this.accountRepository = accountRepository;
        this.transactionRepository = transactionRepository;
        this.ledgerEntryRepository = ledgerEntryRepository;
        this.auditService = auditService;
    }

    /**
     * Standard reducing-balance EMI formula:
     *   EMI = P * r * (1+r)^n / ((1+r)^n - 1)
     * where r is the *monthly* interest rate (annual rate / 12 / 100).
     */
    public LoanCalculatorResponse calculate(LoanCalculatorRequest request) {
        EmiResult result = computeEmi(request.principal(), request.annualInterestRate(), request.tenureMonths());
        return new LoanCalculatorResponse(result.emi(), result.totalInterest(), result.totalRepayment());
    }

    @Transactional
    public LoanResponse apply(LoanApplicationRequest request, Long customerId) {
        Account account = accountRepository.findById(request.accountId())
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));
        if (!account.getCustomer().getId().equals(customerId)) {
            throw new com.corebank.exception.UnauthorizedException("This account does not belong to you");
        }

        EmiResult result = computeEmi(request.principal(), request.annualInterestRate(), request.tenureMonths());

        Loan loan = new Loan();
        loan.setCustomer(account.getCustomer());
        loan.setAccount(account);
        loan.setPrincipalAmount(request.principal());
        loan.setInterestRate(request.annualInterestRate());
        loan.setTenureMonths(request.tenureMonths());
        loan.setEmiAmount(result.emi());
        loan.setTotalRepayment(result.totalRepayment());
        loan.setOutstandingAmount(result.totalRepayment());
        // Simulated instant approval + disbursement, since this is a local
        // simulation rather than a real underwriting pipeline.
        loan.setStatus(Loan.Status.ACTIVE);
        loan = loanRepository.save(loan);

        disburse(loan, account, customerId);

        auditService.log(customerId, "LOAN_APPLICATION", "LOAN", loan.getId(),
                "Applied for and disbursed simulated loan of " + request.principal());

        return LoanResponse.from(loan);
    }

    @Transactional(readOnly = true)
    public List<LoanResponse> list(Long customerId) {
        return loanRepository.findByCustomerId(customerId).stream()
                .map(LoanResponse::from)
                .toList();
    }

    /**
     * Fictional credit score derived from simple, explainable simulated
     * factors. Clearly not a real credit bureau score — see docs/README.
     */
    @Transactional(readOnly = true)
    public CreditScoreResponse creditScore(Long customerId) {
        List<Account> accounts = accountRepository.findByCustomerId(customerId);
        List<Loan> loans = loanRepository.findByCustomerId(customerId);

        int base = 650;
        int accountAgeBonus = Math.min(accounts.size() * 15, 60);
        int activeLoanPenalty = (int) loans.stream().filter(l -> l.getStatus() == Loan.Status.ACTIVE).count() * 20;
        int closedLoanBonus = (int) loans.stream().filter(l -> l.getStatus() == Loan.Status.CLOSED).count() * 25;

        int score = base + accountAgeBonus + closedLoanBonus - activeLoanPenalty;
        score = Math.max(300, Math.min(900, score));

        String rating;
        if (score >= 800) rating = "EXCELLENT";
        else if (score >= 740) rating = "VERY GOOD";
        else if (score >= 670) rating = "GOOD";
        else if (score >= 580) rating = "FAIR";
        else rating = "POOR";

        return CreditScoreResponse.of(score, rating);
    }

    // ------------------------------------------------------------------

    private void disburse(Loan loan, Account account, Long customerId) {
        Account locked = accountRepository.findByIdForUpdate(account.getId())
                .orElseThrow(() -> new AccountNotFoundException("Account not found"));

        BigDecimal newBalance = locked.getBalance().add(loan.getPrincipalAmount());
        locked.setBalance(newBalance);
        locked.setAvailableBalance(locked.getAvailableBalance().add(loan.getPrincipalAmount()));
        accountRepository.save(locked);

        Transaction txn = new Transaction();
        txn.setTransactionRef("TXN" + LocalDateTime.now().format(REF_FORMAT) +
                ThreadLocalRandom.current().nextInt(100000, 999999));
        txn.setType(Transaction.Type.LOAN_DISBURSEMENT);
        txn.setDestinationAccount(locked);
        txn.setAmount(loan.getPrincipalAmount());
        txn.setDescription("Loan disbursement for loan #" + loan.getId());
        txn.setStatus(Transaction.Status.SUCCESS);
        transactionRepository.save(txn);

        LedgerEntry entry = new LedgerEntry();
        entry.setTransaction(txn);
        entry.setAccount(locked);
        entry.setEntryType(LedgerEntry.EntryType.CREDIT);
        entry.setAmount(loan.getPrincipalAmount());
        entry.setBalanceAfter(newBalance);
        ledgerEntryRepository.save(entry);
    }

    private EmiResult computeEmi(BigDecimal principal, BigDecimal annualRate, int tenureMonths) {
        MathContext mc = new MathContext(12);
        BigDecimal monthlyRate = annualRate.divide(BigDecimal.valueOf(1200), mc);

        if (monthlyRate.compareTo(BigDecimal.ZERO) == 0) {
            BigDecimal emi = principal.divide(BigDecimal.valueOf(tenureMonths), 2, RoundingMode.HALF_UP);
            BigDecimal totalRepayment = emi.multiply(BigDecimal.valueOf(tenureMonths));
            return new EmiResult(emi, BigDecimal.ZERO, totalRepayment);
        }

        BigDecimal onePlusR = BigDecimal.ONE.add(monthlyRate);
        BigDecimal onePlusRPowN = onePlusR.pow(tenureMonths, mc);

        BigDecimal numerator = principal.multiply(monthlyRate, mc).multiply(onePlusRPowN, mc);
        BigDecimal denominator = onePlusRPowN.subtract(BigDecimal.ONE, mc);

        BigDecimal emi = numerator.divide(denominator, 2, RoundingMode.HALF_UP);
        BigDecimal totalRepayment = emi.multiply(BigDecimal.valueOf(tenureMonths));
        BigDecimal totalInterest = totalRepayment.subtract(principal);

        return new EmiResult(emi, totalInterest, totalRepayment);
    }

    private record EmiResult(BigDecimal emi, BigDecimal totalInterest, BigDecimal totalRepayment) {}
}
