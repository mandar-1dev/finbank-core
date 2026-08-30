package com.corebank.exception;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.AccessDeniedException;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.validation.FieldError;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import java.util.stream.Collectors;

/**
 * Centralized exception handling. Every handler here returns the same
 * {@link ApiError} shape and never leaks stack traces, SQL, or other
 * internal detail back to the frontend.
 */
@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(AccountNotFoundException.class)
    public ResponseEntity<ApiError> handle(AccountNotFoundException ex) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(CustomerNotFoundException.class)
    public ResponseEntity<ApiError> handle(CustomerNotFoundException ex) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(BeneficiaryNotFoundException.class)
    public ResponseEntity<ApiError> handle(BeneficiaryNotFoundException ex) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(CardNotFoundException.class)
    public ResponseEntity<ApiError> handle(CardNotFoundException ex) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(LoanNotFoundException.class)
    public ResponseEntity<ApiError> handle(LoanNotFoundException ex) {
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(InsufficientBalanceException.class)
    public ResponseEntity<ApiError> handle(InsufficientBalanceException ex) {
        return ResponseEntity.status(HttpStatus.UNPROCESSABLE_ENTITY).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(InvalidTransactionException.class)
    public ResponseEntity<ApiError> handle(InvalidTransactionException ex) {
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(AccountFrozenException.class)
    public ResponseEntity<ApiError> handle(AccountFrozenException ex) {
        return ResponseEntity.status(HttpStatus.UNPROCESSABLE_ENTITY).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(CardFrozenException.class)
    public ResponseEntity<ApiError> handle(CardFrozenException ex) {
        return ResponseEntity.status(HttpStatus.UNPROCESSABLE_ENTITY).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(UnauthorizedException.class)
    public ResponseEntity<ApiError> handle(UnauthorizedException ex) {
        return ResponseEntity.status(HttpStatus.FORBIDDEN).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(AccessDeniedException.class)
    public ResponseEntity<ApiError> handle(AccessDeniedException ex) {
        return ResponseEntity.status(HttpStatus.FORBIDDEN).body(ApiError.of("You are not authorized to perform this action"));
    }

    @ExceptionHandler(InvalidCredentialsException.class)
    public ResponseEntity<ApiError> handle(InvalidCredentialsException ex) {
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(BadCredentialsException.class)
    public ResponseEntity<ApiError> handle(BadCredentialsException ex) {
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(ApiError.of("Invalid email or password"));
    }

    @ExceptionHandler(DuplicateTransactionException.class)
    public ResponseEntity<ApiError> handle(DuplicateTransactionException ex) {
        return ResponseEntity.status(HttpStatus.CONFLICT).body(ApiError.of(ex.getMessage()));
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ApiError> handle(MethodArgumentNotValidException ex) {
        String message = ex.getBindingResult().getFieldErrors().stream()
                .map(FieldError::getDefaultMessage)
                .collect(Collectors.joining("; "));
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(ApiError.of(message.isBlank() ? "Invalid request" : message));
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<ApiError> handle(IllegalArgumentException ex) {
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(ApiError.of(ex.getMessage()));
    }

    /** Last-resort catch-all — never exposes the real exception to the client. */
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ApiError> handle(Exception ex) {
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(ApiError.of("An unexpected error occurred. Please try again."));
    }
}
