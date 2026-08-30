package com.corebank.service;

import com.corebank.entity.AuditLog;
import com.corebank.repository.AuditLogRepository;
import org.springframework.stereotype.Service;

@Service
public class AuditService {

    private final AuditLogRepository auditLogRepository;

    public AuditService(AuditLogRepository auditLogRepository) {
        this.auditLogRepository = auditLogRepository;
    }

    /**
     * Records a sensitive action. Deliberately does not throw on failure
     * paths of the caller — audit writes happen alongside, not instead of,
     * the primary operation's own transaction handling.
     */
    public void log(Long userId, String action, String entityType, Long entityId, String description) {
        AuditLog entry = new AuditLog();
        entry.setUserId(userId);
        entry.setAction(action);
        entry.setEntityType(entityType);
        entry.setEntityId(entityId);
        entry.setDescription(description);
        auditLogRepository.save(entry);
    }
}
