package com.liulimi.fastapi.utils;

import com.liulimi.fastapi.framework.util.TOTPUtils;
import lombok.extern.slf4j.Slf4j;
import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;

@SpringBootTest
@Slf4j
public class TOTPTest {
    @Test
    public void generateSecretKey() {
        String secret = TOTPUtils.generateSecretKey();
        log.info("secret:{}", secret);
    }

    @Test
    public void generateCode() {
        String secret = "6shyg3uens2sh5slhey3dmh47skvgq5y";

        String passcode = TOTPUtils.generateTOTP(secret);
        log.info("passcode:{}", passcode);
    }

    @Test
    public void generateValid() {
        for(int i = 0; i < 50; i++) {
            String secret = TOTPUtils.generateSecretKey();
            String passcode = TOTPUtils.generateTOTP(secret);
            log.info("secret: {}, passcode: {}", secret, passcode);
        }
    }
}
