package com.contful.framework.util;

import cn.hutool.core.date.DateUtil;
import com.contful.framework.config.ProjectConfig;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;
import java.util.Map;

/**
 * 扩展JWT
 *
 * @author boolean 2022-09-15 11:27
 */
@Slf4j
@Component
public final class JWTUtil extends cn.hutool.jwt.JWT {
    @Autowired
    private ProjectConfig projectConfig;

    /**
     * 创建令牌
     *
     * @param payloadClaims 参与令牌对象
     * @return 令牌字符串
     */
    public String createToken(Map<String, ?> payloadClaims) {
        String secret = projectConfig.getSecure().getSecret();
        ProjectConfig.Token token = projectConfig.getToken();

        return create().addPayloads(payloadClaims)
                // 设置签发时间
                .setIssuedAt(DateUtil.date())
                // 设置生效时间
                .setNotBefore(DateUtil.date())
                // 失效时间
                .setExpiresAt(DateUtil.date(DateUtil.current() + token.getExpireTime()))
                .setKey(secret.getBytes(StandardCharsets.UTF_8))
                .sign();
    }
}
