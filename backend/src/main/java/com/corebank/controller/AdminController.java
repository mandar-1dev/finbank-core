package com.corebank.controller;

import com.corebank.dto.AccountResponse;
import com.corebank.dto.AdminDashboardResponse;
import com.corebank.dto.AuditLogResponse;
import com.corebank.dto.CustomerSummaryResponse;
import com.corebank.service.AccountService;
import com.corebank.service.AdminService;
import com.corebank.util.CurrentUser;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/admin")
public class AdminController {

    private final AdminService adminService;
    private final AccountService accountService;

    public AdminController(AdminService adminService, AccountService accountService) {
        this.adminService = adminService;
        this.accountService = accountService;
    }

    @GetMapping("/dashboard")
    public AdminDashboardResponse dashboard() {
        return adminService.dashboard();
    }

    @PostMapping("/accounts/{id}/freeze")
    public void freeze(@PathVariable Long id, Authentication authentication) {
        accountService.freezeAccount(id, CurrentUser.id(authentication));
    }

    @PostMapping("/accounts/{id}/unfreeze")
    public void unfreeze(@PathVariable Long id, Authentication authentication) {
        accountService.unfreezeAccount(id, CurrentUser.id(authentication));
    }

    @GetMapping("/audit-logs")
    public Page<AuditLogResponse> auditLogs(@RequestParam(defaultValue = "0") int page,
                                             @RequestParam(defaultValue = "20") int size) {
        return adminService.auditLogs(PageRequest.of(page, size));
    }

    @GetMapping("/customers")
    public Page<CustomerSummaryResponse> searchCustomers(@RequestParam(required = false) String query,
                                                          @RequestParam(defaultValue = "0") int page,
                                                          @RequestParam(defaultValue = "20") int size) {
        return adminService.searchCustomers(query, PageRequest.of(page, size));
    }

    @GetMapping("/customers/{id}/accounts")
    public List<AccountResponse> customerAccounts(@PathVariable Long id) {
        return adminService.customerAccounts(id);
    }
}
