package com.corebank.controller;

import com.corebank.dto.*;
import com.corebank.service.CardService;
import com.corebank.util.CurrentUser;
import jakarta.validation.Valid;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/cards")
public class CardController {

    private final CardService cardService;

    public CardController(CardService cardService) {
        this.cardService = cardService;
    }

    @GetMapping
    public List<CardResponse> list(Authentication authentication) {
        return cardService.list(CurrentUser.id(authentication));
    }

    @PostMapping("/{id}/freeze")
    public CardResponse freeze(@PathVariable Long id, Authentication authentication) {
        return cardService.freeze(id, CurrentUser.id(authentication));
    }

    @PostMapping("/{id}/unfreeze")
    public CardResponse unfreeze(@PathVariable Long id, Authentication authentication) {
        return cardService.unfreeze(id, CurrentUser.id(authentication));
    }

    @PutMapping("/{id}/limit")
    public CardResponse setLimit(@PathVariable Long id, @Valid @RequestBody SpendingLimitRequest request,
                                  Authentication authentication) {
        return cardService.setSpendingLimit(id, request, CurrentUser.id(authentication));
    }

    @PostMapping("/pay")
    public TransactionResponse pay(@Valid @RequestBody CardPaymentRequest request, Authentication authentication) {
        return cardService.pay(request, CurrentUser.id(authentication));
    }
}
