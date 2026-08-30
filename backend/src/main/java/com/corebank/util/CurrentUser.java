package com.corebank.util;

import com.corebank.security.CustomerPrincipal;
import org.springframework.security.core.Authentication;

/** Small helper so controllers don't repeat the principal-casting boilerplate. */
public final class CurrentUser {

    private CurrentUser() {}

    public static Long id(Authentication authentication) {
        return ((CustomerPrincipal) authentication.getPrincipal()).getId();
    }

    public static CustomerPrincipal principal(Authentication authentication) {
        return (CustomerPrincipal) authentication.getPrincipal();
    }
}
