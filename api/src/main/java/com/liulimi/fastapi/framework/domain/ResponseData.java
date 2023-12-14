package com.liulimi.fastapi.framework.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonInclude.Include;
import lombok.Data;
import org.springframework.http.HttpStatus;

import java.io.Serial;
import java.io.Serializable;

/**
 * ResponseData 封装
 *
 * @author Boolean 2021-04-19 23:40
 */
@Data
@SuppressWarnings("unchecked")
@JsonInclude(Include.NON_NULL)
public class ResponseData<T> implements Serializable {
    @Serial
    private static final long serialVersionUID = 1L;
    private Integer code;
    private String message;
    private Integer timestamp;
    private T data;

    public ResponseData(Integer code, String message, T data) {
        this.code = code;
        this.message = message;
        this.data = data;
        this.timestamp = (int) (System.currentTimeMillis() / 1000);
    }

    public static ResponseData success() {
        return new ResponseData(HttpStatus.OK.value(), "success", null);
    }

    public static ResponseData success(String message) {
        return new ResponseData(HttpStatus.OK.value(), message, null);
    }

    public static ResponseData success(Object data) {
        return new ResponseData(HttpStatus.OK.value(), "success", data);
    }

    public static <T> ResponseData success(Integer code, T data) {
        return new ResponseData(code, "success", data);
    }

    public static <T> ResponseData success(HttpStatus httpStatus, T data) {
        return new ResponseData(httpStatus.value(), "success", data);
    }

    public static <T> ResponseData success(String message, T data) {
        return new ResponseData(HttpStatus.OK.value(), message, data);
    }

    public static ResponseData fail() {
        return new ResponseData(HttpStatus.INTERNAL_SERVER_ERROR.value(), "failed", null);
    }

    public static ResponseData fail(String message) {
        return new ResponseData(HttpStatus.INTERNAL_SERVER_ERROR.value(), message, null);
    }

    public static ResponseData fail(Object data) {
        return new ResponseData(HttpStatus.INTERNAL_SERVER_ERROR.value(), "failed", data);
    }

    public static ResponseData fail(Integer code, String message) {
        return new ResponseData(code, message, null);
    }

    public static ResponseData fail(HttpStatus httpStatus, String message) {
        return new ResponseData(httpStatus.value(), message, null);
    }

    public static <T> ResponseData fail(String message, T data) {
        return new ResponseData(HttpStatus.NOT_IMPLEMENTED.value(), message, data);
    }

    public static <T> ResponseData build(Integer code, String message, T data) {
        return new ResponseData(code, message, data);
    }
}
