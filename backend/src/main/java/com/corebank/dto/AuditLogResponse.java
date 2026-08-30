package com.corebank.dto;

import com.corebank.entity.AuditLog;

import java.time.LocalDateTime;

public record AuditLogResponse(
        Long id,
        Long userId,
        String action,
        String entityType,
        Long entityId,
        String description,
        LocalDateTime timestamp
) {
    public static AuditLogResponse from(AuditLog log) {
        return new AuditLogResponse(log.getId(), log.getUserId(), log.getAction(),
                log.getEntityType(), log.getEntityId(), log.getDescription(), log.getTimestamp());
    }
}
