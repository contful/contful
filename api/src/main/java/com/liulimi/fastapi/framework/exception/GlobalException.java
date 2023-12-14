package com.liulimi.fastapi.framework.exception;

import com.liulimi.fastapi.framework.annotation.AuthIgnore;
import com.liulimi.fastapi.framework.domain.ResponseData;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.web.servlet.error.ErrorController;
import org.springframework.http.HttpStatus;
import org.springframework.validation.BindException;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.NoHandlerFoundException;

/**
 * 自定义全局异常处理类
 *
 * @author boolean
 */
@RestController
@RestControllerAdvice
public class GlobalException implements ErrorController {
    private final Logger logger = LoggerFactory.getLogger(this.getClass());

    /**
     * 不存在的请求路径
     *
     * @return NOT_FOUND
     */
    @AuthIgnore
    @RequestMapping("/error")
    @ResponseStatus(HttpStatus.NOT_FOUND)
    public Object notFoundHandler() {
        return ResponseData.fail(HttpStatus.NOT_FOUND, "Page not found");
    }

    /**
     * 基础异常
     *
     * @param e 接收异常信息参数
     * @return Object
     */
    @ExceptionHandler(BaseException.class)
    public Object baseException(BaseException e) {
        logger.error("发生业务异常！原因是：{}", e.getMessage(), e);
        return ResponseData.fail(e.getCode(), e.getMessage());
    }

    /**
     * 处理其他异常
     *
     * @param e 接收异常信息参数
     * @return Object
     */
    @ExceptionHandler(Exception.class)
    public Object handleException(Exception e) {
        logger.error("发生系统异常！原因是：{}", e.getMessage(), e);
        return ResponseData.fail(e.getMessage());
    }

    /**
     * 自定义验证异常
     *
     * @param e 接收异常信息参数
     * @return Object
     */
    @ExceptionHandler(BindException.class)
    public Object validatedBindException(BindException e) {
        logger.error(e.getMessage(), e);
        return ResponseData.fail(e.getMessage());
    }

    /**
     * 自定义验证异常
     *
     * @param e 接收异常信息参数
     * @return Object
     */
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public Object validExceptionHandler(MethodArgumentNotValidException e) {
        logger.error(e.getMessage(), e);
        return ResponseData.fail(e.getMessage());
    }

    /**
     * 处理空指针的异常
     *
     * @param e 接收异常信息参数
     * @return Object
     */
    @ExceptionHandler(NullPointerException.class)
    public Object handlerNullPointerException(Exception e) {
        logger.error("发生空指针异常！原因是:", e.getMessage(), e);
        return ResponseData.fail(HttpStatus.INTERNAL_SERVER_ERROR, "发生空指针异常！");
    }

    /**
     * 路径不存在
     *
     * @param e 接收异常信息参数
     * @return object 对象
     */
    @ExceptionHandler(NoHandlerFoundException.class)
    public Object handlerNoFoundException(Exception e) {
        logger.error("路径不存在，请检查路径是否正确:", e.getMessage(), e);
        return ResponseData.fail(HttpStatus.NOT_FOUND, "路径不存在，请检查路径是否正确");
    }
}
