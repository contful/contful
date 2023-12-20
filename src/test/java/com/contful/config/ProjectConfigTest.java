package com.contful.config;

import com.contful.framework.config.ProjectConfig;
import lombok.extern.slf4j.Slf4j;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

@SpringBootTest
@Slf4j
public class ProjectConfigTest {
    @Autowired
    ProjectConfig projectConfig;

    @Test
    public void config() {
        String name = projectConfig.getName();
        String version = projectConfig.getVersion();
        String xss = projectConfig.getSecure().getXss().getUrlPatterns();
        Integer tokenExpire = projectConfig.getToken().getExpireTime();
        log.info("name:{},vsersion:{},xss:{},tokenExpire:{}", name, version, xss, tokenExpire);
    }
}
