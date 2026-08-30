package com.corebank.service;

import com.corebank.dto.*;
import com.corebank.entity.Account;
import com.corebank.entity.Card;
import com.corebank.entity.LedgerEntry;
import com.corebank.entity.Transaction;
import com.corebank.exception.*;
import com.corebank.repository.AccountRepository;
import com.corebank.repository.CardRepository;
import com.corebank.repository.LedgerEntryRepository;
import com.corebank.repository.TransactionRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.concurrent.ThreadLocalRandom;

@Service
public class CardService {

    private static final DateTimeFormatter REF_FORMAT = DateTimeFormatter.ofPattern("yyyyMMdd");

    private final CardRepository cardRepository;
    private final AccountRepository accountRepository;
    private final TransactionRepository transactionRepository;
    private final LedgerEntryRepository ledgerEntryRepository;
    private final AuditService auditService;

    public CardService(CardRepository cardRepository, AccountRepository accountRepository,
                        TransactionRepository transactionRepository, LedgerEntryRepository ledgerEntryRepository,
                        AuditService auditService) {
        this.cardRepository = cardRepository;
        this.accountRepository = accountRepository;
        this.transactionRepository = transactionRepository;
        this.ledgerEntryRepository = ledgerEntryRepository;
        this.auditService = auditService;
    }

    @Transactional(readOnly = true)
    public List<CardResponse> list(Long customerId) {
        return cardRepository.findByAccountCustomerId(customerId).stream()
                .map(CardResponse::from)
                .toList();
    }

    @Transactional
    public CardResponse freeze(Long cardId, Long customerId) {
        Card card = ownedCard(cardId, customerId);
        card.setStatus(Card.Status.FROZEN);
        cardRepository.save(card);
        auditService.log(customerId, "CARD_FREEZE", "CARD", cardId, "Card frozen by customer");
        return CardResponse.from(card);
    }

    @Transactional
    public CardResponse unfreeze(Long cardId, Long customerId) {
        Card card = ownedCard(cardId, customerId);
        card.setStatus(Card.Status.ACTIVE);
        cardRepository.save(card);
        auditService.log(customerId, "CARD_UNFREEZE", "CARD", cardId, "Card unfrozen by customer");
        return CardResponse.from(card);
    }

    @Transactional
    public CardResponse setSpendingLimit(Long cardId, SpendingLimitRequest request, Long customerId) {
        Card card = ownedCard(cardId, customerId);
        card.setSpendingLimit(request.spendingLimit());
        cardRepository.save(card);
        auditService.log(customerId, "CARD_LIMIT_UPDATE", "CARD", cardId,
                "Spending limit updated to " + request.spendingLimit());
        return CardResponse.from(card);
    }

    /**
     * Simulated card payment. Validates the card, checks the per-card
     * spending limit and the account balance, then debits the linked
     * account inside the same locked, transactional flow used for
     * transfers.
     */
    @Transactional
    public TransactionResponse pay(CardPaymentRequest request, Long customerId) {
        Card card = ownedCard(request.cardId(), customerId);

        if (card.getStatus() == Card.Status.FROZEN) {
            throw new CardFrozenException("Card is frozen. Unfreeze it before making a payment.");
        }
        if (card.getStatus() != Card.Status.ACTIVE) {
            throw new InvalidTransactionException("Card is not active");
        }
        if (request.amount().compareTo(card.getSpendingLimit()) > 0) {
            throw new InvalidTransactionException("Amount exceeds the card's spending limit");
        }

        Account account = accountRepository.findByIdForUpdate(card.getAccount().getId())
                .orElseThrow(() -> new AccountNotFoundException("Linked account not found"));
        if (account.getStatus() != Account.Status.ACTIVE) {
            throw new AccountFrozenException("Linked account is not active");
        }
        if (account.getAvailableBalance().compareTo(request.amount()) < 0) {
            throw new InsufficientBalanceException("Insufficient account balance");
        }

        BigDecimal newBalance = account.getBalance().subtract(request.amount());
        account.setBalance(newBalance);
        account.setAvailableBalance(account.getAvailableBalance().subtract(request.amount()));
        accountRepository.save(account);

        Transaction txn = new Transaction();
        txn.setTransactionRef("TXN" + LocalDateTime.now().format(REF_FORMAT) +
                ThreadLocalRandom.current().nextInt(100000, 999999));
        txn.setType(Transaction.Type.CARD_PAYMENT);
        txn.setSourceAccount(account);
        txn.setAmount(request.amount());
        txn.setDescription("Card payment to " + request.merchant());
        txn.setStatus(Transaction.Status.SUCCESS);
        transactionRepository.save(txn);

        LedgerEntry entry = new LedgerEntry();
        entry.setTransaction(txn);
        entry.setAccount(account);
        entry.setEntryType(LedgerEntry.EntryType.DEBIT);
        entry.setAmount(request.amount());
        entry.setBalanceAfter(newBalance);
        ledgerEntryRepository.save(entry);

        auditService.log(customerId, "CARD_PAYMENT", "TRANSACTION", txn.getId(),
                "Card payment of " + request.amount() + " to " + request.merchant());

        return TransactionResponse.from(txn);
    }

    private Card ownedCard(Long cardId, Long customerId) {
        return cardRepository.findByIdAndAccountCustomerId(cardId, customerId)
                .orElseThrow(() -> new CardNotFoundException("Card not found"));
    }
}
