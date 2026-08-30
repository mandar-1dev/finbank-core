package com.corebank.dto;

/** Generic success envelope: { "success": true, "message": "...", "data": ... } */
public record ApiResponse<T>(boolean success, String message, T data) {
    public static <T> ApiResponse<T> of(String message, T data) {
        return new ApiResponse<>(true, message, data);
    }
    public static ApiResponse<Void> of(String message) {
        return new ApiResponse<>(true, message, null);
    }
}
