package com.corebank.controller;

import com.corebank.dto.*;
import com.corebank.service.LoanService;
import com.corebank.util.CurrentUser;
import jakarta.validation.Valid;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/loans")
public class LoanController {

    private final LoanService loanService;

    public LoanController(LoanService loanService) {
        this.loanService = loanService;
    }

    @PostMapping("/calculate")
    public LoanCalculatorResponse calculate(@Valid @RequestBody LoanCalculatorRequest request) {
        return loanService.calculate(request);
    }

    @PostMapping("/apply")
    public LoanResponse apply(@Valid @RequestBody LoanApplicationRequest request, Authentication authentication) {
        return loanService.apply(request, CurrentUser.id(authentication));
    }

    @GetMapping
    public List<LoanResponse> list(Authentication authentication) {
        return loanService.list(CurrentUser.id(authentication));
    }

    @GetMapping("/credit-score")
    public CreditScoreResponse creditScore(Authentication authentication) {
        return loanService.creditScore(CurrentUser.id(authentication));
    }
}
