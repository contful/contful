package com.contful.framework.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;
import org.springframework.stereotype.Component;

@Data
@Component
@Configuration
@ConfigurationProperties(prefix = "project", ignoreUnknownFields = true)
public class ProjectConfig {
    /**
     * 项目名称
     */
    private String name = "FastAPI";
    /**
     * 版本
     */
    private String version = "1.1.0";
    /**
     * 安全
     */
    private Secure secure;
    /**
     * Token
     */
    private Token token;

    @Data
    public static class Secure {
        /**
         * 参数签名
         */
        private Sign sign;
        /**
         * 安全密钥
         */
        private String secret = "abcdefghijklmnopqrstuvwxyz";
        private Xss xss;

        @Data
        public static class Sign {
            private Integer maxTime;
        }

        @Data
        public static class Xss {
            private String enabled;
            private String excludes;
            private String urlPatterns;
        }
    }

    @Data
    public static class Token {
        /**
         * Request Headers ：Authorization
         */
        private String header = "Authorization";

        /**
         * 令牌过期时间 此处单位/秒
         */
        private Integer expireTime = 1800;
    }
}
