package com.corebank.exception;

import com.fasterxml.jackson.annotation.JsonFormat;

import java.time.LocalDateTime;

/**
 * Uniform error shape returned by every API failure. Never carries a stack
 * trace, SQL text, or any other internal detail — only a safe message.
 */
public record ApiError(
        boolean success,
        String message,
        @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss") LocalDateTime timestamp
) {
    public static ApiError of(String message) {
        return new ApiError(false, message, LocalDateTime.now());
    }
}
