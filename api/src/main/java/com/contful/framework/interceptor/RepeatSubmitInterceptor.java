package com.contful.framework.interceptor;

import com.contful.framework.annotation.RepeatSubmit;
import com.contful.framework.exception.BaseException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Component;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.HandlerInterceptor;

import java.lang.reflect.Method;

/**
 * 防止重复提交拦截器
 *
 * @author boolean 2021-05-11 23:04
 */
@Component
public abstract class RepeatSubmitInterceptor implements HandlerInterceptor {
    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        if (!(handler instanceof HandlerMethod)) {
            return true;
        }

        Method method = ((HandlerMethod) handler).getMethod();
        RepeatSubmit annotation = method.getAnnotation(RepeatSubmit.class);
        if (annotation == null) {
            return true;
        }

        if (this.isRepeatSubmit(request)) {
            throw new BaseException(HttpStatus.NOT_ACCEPTABLE.value(), "不允许重复提交，请稍后再试");
        }

        return true;
    }

    /**
     * 验证是否重复提交由子类实现具体地防重复提交的规则
     *
     * @param request http请求
     * @return boolean
     */
    public abstract boolean isRepeatSubmit(HttpServletRequest request);
}
