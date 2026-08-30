package com.corebank.service;

import com.corebank.dto.LoginRequest;
import com.corebank.dto.LoginResponse;
import com.corebank.entity.Customer;
import com.corebank.exception.InvalidCredentialsException;
import com.corebank.repository.CustomerRepository;
import com.corebank.security.CustomerPrincipal;
import com.corebank.security.JwtService;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.DisabledException;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class AuthService {

    private final AuthenticationManager authenticationManager;
    private final CustomerRepository customerRepository;
    private final JwtService jwtService;
    private final AuditService auditService;

    public AuthService(AuthenticationManager authenticationManager, CustomerRepository customerRepository,
                        JwtService jwtService, AuditService auditService) {
        this.authenticationManager = authenticationManager;
        this.customerRepository = customerRepository;
        this.jwtService = jwtService;
        this.auditService = auditService;
    }

    @Transactional
    public LoginResponse login(LoginRequest request) {
        try {
            authenticationManager.authenticate(
                    new UsernamePasswordAuthenticationToken(request.email(), request.password()));
        } catch (DisabledException ex) {
            throw new InvalidCredentialsException("Your account is suspended. Please contact support.");
        } catch (BadCredentialsException ex) {
            throw new InvalidCredentialsException("Invalid email or password");
        }

        Customer customer = customerRepository.findByEmail(request.email())
                .orElseThrow(() -> new InvalidCredentialsException("Invalid email or password"));

        String token = jwtService.generateToken(new CustomerPrincipal(customer));
        auditService.log(customer.getId(), "LOGIN", "CUSTOMER", customer.getId(), "Successful login");

        return LoginResponse.of(token, customer.getId(), customer.getFullName(), customer.getRole().name());
    }
}
