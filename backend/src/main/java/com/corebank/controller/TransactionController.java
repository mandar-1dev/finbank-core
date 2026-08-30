package com.corebank.controller;

import com.corebank.dto.*;
import com.corebank.entity.Transaction;
import com.corebank.service.TransactionService;
import com.corebank.util.CurrentUser;
import jakarta.validation.Valid;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;

@RestController
@RequestMapping("/api/transactions")
public class TransactionController {

    private final TransactionService transactionService;

    public TransactionController(TransactionService transactionService) {
        this.transactionService = transactionService;
    }

    @PostMapping("/deposit")
    public TransactionResponse deposit(@Valid @RequestBody DepositRequest request, Authentication authentication) {
        return transactionService.deposit(request, CurrentUser.id(authentication));
    }

    @PostMapping("/withdraw")
    public TransactionResponse withdraw(@Valid @RequestBody WithdrawRequest request, Authentication authentication) {
        return transactionService.withdraw(request, CurrentUser.id(authentication));
    }

    /**
     * Transfer supports idempotency two ways: an "Idempotency-Key" header
     * (standard payment-API convention) or a key embedded in the request
     * body — whichever is supplied is honored.
     */
    @PostMapping("/transfer")
    public TransactionResponse transfer(@Valid @RequestBody TransferRequest request,
                                         @RequestHeader(value = "Idempotency-Key", required = false) String headerKey,
                                         Authentication authentication) {
        String key = headerKey != null ? headerKey : request.idempotencyKey();
        TransferRequest effective = new TransferRequest(request.sourceAccountId(), request.destinationAccountNumber(),
                request.beneficiaryId(), request.amount(), request.description(), key);
        return transactionService.transfer(effective, CurrentUser.id(authentication));
    }

    @GetMapping
    public Page<TransactionResponse> history(@RequestParam Long accountId,
                                              @RequestParam(required = false) Transaction.Type type,
                                              @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime from,
                                              @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime to,
                                              @RequestParam(defaultValue = "0") int page,
                                              @RequestParam(defaultValue = "20") int size,
                                              Authentication authentication) {
        return transactionService.getHistory(accountId, CurrentUser.id(authentication), type, from, to,
                PageRequest.of(page, size));
    }
}
