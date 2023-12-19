package com.contful.cache;

import com.contful.framework.util.RedisUtil;
import lombok.extern.slf4j.Slf4j;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

@Slf4j
@SpringBootTest
public class RedisTest {
    @Autowired
    private RedisUtil redisUtil;

    @Test
    public void redisUtil() {
        String cacheKey = "boot:test:redis:util:key";
        redisUtil.set(cacheKey, "hello redis util", 100);
        log.info("Key:{},Value:{}", cacheKey, redisUtil.get(cacheKey));
    }
}