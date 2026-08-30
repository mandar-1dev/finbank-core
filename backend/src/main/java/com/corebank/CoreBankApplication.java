package com.corebank;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * CoreBank — Core Banking Simulation Platform.
 *
 * Educational project. Simulates customers, accounts, transfers, cards
 * and loans against a local MySQL database. Never connects to a real
 * bank, payment network, or financial institution.
 */
@SpringBootApplication
public class CoreBankApplication {
    public static void main(String[] args) {
        SpringApplication.run(CoreBankApplication.class, args);
    }
}
