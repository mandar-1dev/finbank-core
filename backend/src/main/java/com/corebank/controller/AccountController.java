package com.corebank.controller;

import com.corebank.dto.AccountResponse;
import com.corebank.service.AccountService;
import com.corebank.util.CurrentUser;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/accounts")
public class AccountController {

    private final AccountService accountService;

    public AccountController(AccountService accountService) {
        this.accountService = accountService;
    }

    @GetMapping
    public List<AccountResponse> myAccounts(Authentication authentication) {
        return accountService.getAccountsForCustomer(CurrentUser.id(authentication));
    }

    @GetMapping("/{id}")
    public AccountResponse getAccount(@PathVariable Long id, Authentication authentication) {
        return accountService.getAccountForCustomer(id, CurrentUser.id(authentication));
    }
}
