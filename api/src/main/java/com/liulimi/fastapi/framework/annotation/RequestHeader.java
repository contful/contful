package com.liulimi.fastapi.framework.annotation;

import com.liulimi.fastapi.framework.exception.BaseException;
import lombok.Getter;
import org.apache.commons.lang3.ObjectUtils;
import org.apache.commons.lang3.StringUtils;
import org.springframework.http.HttpStatus;

/**
 * @author boolean
 * 基于nonce + timestamp 方案，实现接口的防篡改和防重放
 * @date 2023-02-17 11:12
 */
@Getter
public class RequestHeader {
    /**
     * 加密段
     */
    private final String sign;
    /**
     * 本次请求时间戳
     */
    private final Long timestamp;
    /**
     * 请求随机字符串
     */
    private final String nonce;

    private RequestHeader(Builder builder) {
        // 任何信息不能为空
        if (StringUtils.isEmpty(builder.sign) ||
                ObjectUtils.isEmpty(builder.timestamp) ||
                StringUtils.isEmpty(builder.nonce)) {
            throw new BaseException(HttpStatus.BAD_REQUEST.value(), "拒绝非法请求标头");
        }

        sign = builder.sign;
        timestamp = builder.timestamp;
        nonce = builder.nonce;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static final class Builder {
        private String sign;
        private Long timestamp;
        private String nonce;

        private Builder() {
        }

        public Builder sign(String sign) {
            this.sign = sign;
            return this;
        }

        public Builder timestamp(Long timestamp) {
            this.timestamp = timestamp;
            return this;
        }

        public Builder nonce(String nonce) {
            this.nonce = nonce;
            return this;
        }

        public RequestHeader build() {
            return new RequestHeader(this);
        }
    }
}
