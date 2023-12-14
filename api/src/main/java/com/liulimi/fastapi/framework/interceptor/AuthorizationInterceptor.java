package com.liulimi.fastapi.framework.interceptor;

import cn.hutool.core.date.DateUtil;
import cn.hutool.jwt.JWT;
import cn.hutool.jwt.JWTValidator;
import com.liulimi.fastapi.framework.annotation.AuthIgnore;
import com.liulimi.fastapi.framework.config.ProjectConfig;
import com.liulimi.fastapi.framework.exception.BaseException;
import com.liulimi.fastapi.framework.util.StringUtil;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Component;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.HandlerInterceptor;
import org.springframework.web.servlet.ModelAndView;

import java.nio.charset.StandardCharsets;
import java.util.Optional;

/**
 * 认证拦截器
 *
 * @author boolean 2022-08-17 09:51
 */
@Component
public class AuthorizationInterceptor implements HandlerInterceptor {

    @Autowired
    private ProjectConfig projectConfig;

    /**
     * 目标方法执行之前
     *
     * @param request  请求参数
     * @param response 响应内容
     * @param handler  句柄对象
     * @return boolean
     */
    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        AuthIgnore annotation;

        // 如果不是映射到方法直接通过
        if (handler instanceof HandlerMethod handlerMethod) {
            annotation = handlerMethod.getMethodAnnotation(AuthIgnore.class);
        } else {
            return true;
        }

        // 检查是否有认证忽略注释，有则跳过认证
        if (annotation != null) {
            return true;
        }

        // 从 http 请求头中取出 token 验证令牌
        try {
            String header = projectConfig.getToken().getHeader();
            String token = request.getHeader(header);
            if (StringUtil.isEmpty(token) || StringUtil.isBlank(token)) {
                throw new BaseException(HttpStatus.BAD_REQUEST.value(), "非法请求，缺少签名");
            }

            JWT jwt = JWT.of(token); // 解析token
            // 验证JWT Token的签名是否有效
            String secret = projectConfig.getSecure().getSecret();
            if (!jwt.setKey(secret.getBytes(StandardCharsets.UTF_8)).verify()) {
                throw new BaseException(HttpStatus.FORBIDDEN.value(), "请求失败，无效签名");
            }

            // 验证时间
            JWTValidator.of(jwt).validateDate(DateUtil.date());
        } catch (Exception e) {
            throw new BaseException(HttpStatus.UNAUTHORIZED.value(), e.getMessage());
        }

        return true;
    }

    /**
     * 目标方法执行完成之后
     *
     * @param request      请求参数
     * @param response     响应内容
     * @param handler      句柄
     * @param modelAndView 模型
     */
    @Override
    public void postHandle(HttpServletRequest request,
                           HttpServletResponse response,
                           Object handler, ModelAndView modelAndView) {

    }

    /**
     * 渲染数据
     *
     * @param request  请求参数
     * @param response 响应内容
     * @param handler  句柄对象
     * @param e        异常信息
     */
    @Override
    public void afterCompletion(HttpServletRequest request,
                                HttpServletResponse response,
                                Object handler, Exception e) {
    }
}
