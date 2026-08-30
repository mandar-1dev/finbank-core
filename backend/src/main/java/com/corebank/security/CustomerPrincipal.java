package com.corebank.security;

import com.corebank.entity.Customer;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.userdetails.UserDetails;

import java.util.Collection;
import java.util.List;

/**
 * Adapts our {@link Customer} entity to Spring Security's UserDetails
 * contract. Roles are exposed as "ROLE_CUSTOMER" / "ROLE_ADMIN" so that
 * @PreAuthorize("hasRole('ADMIN')") works out of the box.
 */
public class CustomerPrincipal implements UserDetails {

    private final Customer customer;

    public CustomerPrincipal(Customer customer) {
        this.customer = customer;
    }

    public Customer getCustomer() {
        return customer;
    }

    public Long getId() {
        return customer.getId();
    }

    @Override
    public Collection<? extends GrantedAuthority> getAuthorities() {
        return List.of(new SimpleGrantedAuthority("ROLE_" + customer.getRole().name()));
    }

    @Override
    public String getPassword() {
        return customer.getPasswordHash();
    }

    @Override
    public String getUsername() {
        return customer.getEmail();
    }

    @Override
    public boolean isAccountNonExpired() { return true; }

    @Override
    public boolean isAccountNonLocked() {
        return customer.getStatus() == Customer.CustomerStatus.ACTIVE;
    }

    @Override
    public boolean isCredentialsNonExpired() { return true; }

    @Override
    public boolean isEnabled() {
        return customer.getStatus() == Customer.CustomerStatus.ACTIVE;
    }
}
