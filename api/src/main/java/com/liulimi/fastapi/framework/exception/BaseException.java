package com.liulimi.fastapi.framework.exception;

import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;

import java.io.Serial;

/**
 * 基础异常
 */
@EqualsAndHashCode(callSuper = true)
@Slf4j
@Data
public class BaseException extends RuntimeException {
    @Serial
    private static final long serialVersionUID = 1L;

    private Integer code = 500;
    private String message;

    public BaseException(String message) {
        super(message);
        this.message = message;
    }

    public BaseException(String message, Throwable e) {
        super(message, e);
        this.message = message;
    }

    public BaseException(Integer code, String message) {
        super(message);
        this.code = code;
        this.message = message;
    }

    public BaseException(HttpStatus httpStatus, String message) {
        super(message);
        this.code = httpStatus.value();
        this.message = message;
    }

    public BaseException(Integer code, String message, Throwable e) {
        super(message, e);
        this.code = code;
        this.message = message;
    }
}
