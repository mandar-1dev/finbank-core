package com.corebank.exception;

public class CustomerNotFoundException extends RuntimeException {
    public CustomerNotFoundException(String message) { super(message); }
}
