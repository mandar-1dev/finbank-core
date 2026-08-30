package com.corebank.exception;

public class CardFrozenException extends RuntimeException {
    public CardFrozenException(String message) { super(message); }
}
