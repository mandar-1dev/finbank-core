package com.corebank.exception;

public class InvalidTransactionException extends RuntimeException {
    public InvalidTransactionException(String message) { super(message); }
}
