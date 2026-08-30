package com.corebank.controller;

import com.corebank.dto.BeneficiaryRequest;
import com.corebank.dto.BeneficiaryResponse;
import com.corebank.service.BeneficiaryService;
import com.corebank.util.CurrentUser;
import jakarta.validation.Valid;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/beneficiaries")
public class BeneficiaryController {

    private final BeneficiaryService beneficiaryService;

    public BeneficiaryController(BeneficiaryService beneficiaryService) {
        this.beneficiaryService = beneficiaryService;
    }

    @GetMapping
    public List<BeneficiaryResponse> list(Authentication authentication) {
        return beneficiaryService.list(CurrentUser.id(authentication));
    }

    @PostMapping
    public BeneficiaryResponse add(@Valid @RequestBody BeneficiaryRequest request, Authentication authentication) {
        return beneficiaryService.add(request, CurrentUser.id(authentication));
    }

    @DeleteMapping("/{id}")
    public void remove(@PathVariable Long id, Authentication authentication) {
        beneficiaryService.remove(id, CurrentUser.id(authentication));
    }
}
