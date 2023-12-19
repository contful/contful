package com.contful.framework.interceptor;

import com.contful.framework.annotation.RequestHeader;
import com.contful.framework.annotation.SignIgnore;
import com.contful.framework.exception.BaseException;
import com.contful.framework.filter.SignRequestWrapper;
import com.contful.framework.util.HttpDataUtil;
import com.contful.framework.util.RedisUtil;
import com.contful.framework.util.SignUtil;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletRequestWrapper;
import jakarta.servlet.http.HttpServletResponse;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Component;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.HandlerInterceptor;

import java.io.IOException;
import java.lang.reflect.Method;
import java.util.SortedMap;

@Slf4j
@Component
public class SignInterceptor implements HandlerInterceptor {
    public static final String HEADER_NONCE = "X-Nonce";
    public static final String HEADER_TIME = "X-Time";
    public static final String HEADER_SIGN = "X-Sign";
    private static final String NONCE_KEY = "x-nonce-";
    @Autowired
    private RedisUtil redisUtil;

    /**
     * 签名超时时间
     */
    @Value("${signMaxTime:60}")
    private Long signMaxTime;

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        // 判断是否需要忽略该请求
        if (isIgnore(handler)) {
            return true;
        }
        log.info("过滤URL:{}", request.getRequestURI());

        HttpServletRequestWrapper requestWrapper = new SignRequestWrapper(request);
        String timeString = request.getHeader(HEADER_TIME);
        if (StringUtils.isEmpty(timeString)) {
            throw new BaseException(HttpStatus.BAD_REQUEST.value(), "非法参数");
        }
        //构建请求头
        RequestHeader requestHeader = RequestHeader.builder()
                .nonce(request.getHeader(HEADER_NONCE))
                .timestamp(Long.parseLong(timeString))
                .sign(request.getHeader(HEADER_SIGN))
                .build();

        // 超时验证
        verifyTimeOut(requestHeader);
        // 验证参数签名
        if (!verifySign(request, requestWrapper, requestHeader)) {
            throw new BaseException(HttpStatus.BAD_REQUEST.value(), "参数签名有误");
        }

        // 重放判定
        verifyReplay(requestHeader);
        return true;
    }

    /**
     * 判断当前请求是否需要忽略
     *
     * @param handler 处理器对象
     * @return 是否需要忽略
     */
    private boolean isIgnore(Object handler) {
        if (!(handler instanceof HandlerMethod handlerMethod)) {
            // 不是Controller方法，直接通过
            return true;
        }

        Class<?> controllerClass = handlerMethod.getBeanType();

        // 判断 Controller 类是否被打上了 SignIgnore 注解
        if (controllerClass.isAnnotationPresent(SignIgnore.class)) {
            return true;
        }

        // 判断当前方法是否被打上了 SignIgnore 注解
        Method method = handlerMethod.getMethod();
        return method.isAnnotationPresent(SignIgnore.class);
    }


    /**
     * 判断timestamp时间戳与当前时间是否操过60s（过期时间根据业务情况设置）,如果超过了就提示签名过期。
     *
     * @param requestHeader 请求头信息
     */
    private void verifyTimeOut(RequestHeader requestHeader) {
        long now = System.currentTimeMillis() / 1000;
        long differTime = now - requestHeader.getTimestamp();
        // 时间取绝对值，即如果传来了未来的时间，在1分钟内也可通过。主要为了兼容时钟不一致的问题
        if (Math.abs(differTime) > signMaxTime) {
            throw new BaseException(HttpStatus.BAD_REQUEST.value(), "请求时间错误");
        }
    }


    /**
     * 判断 nonce 是否已存在
     */
    private void verifyReplay(RequestHeader requestHeader) {
        boolean nonceExists = redisUtil.hasKey(NONCE_KEY + requestHeader.getSign());
        // 请求重复
        if (nonceExists) {
            throw new BaseException(HttpStatus.BAD_REQUEST.value(), "拒绝重复请求");
        } else {
            redisUtil.set(NONCE_KEY + requestHeader.getSign(), requestHeader.getSign(), signMaxTime);
        }
    }

    /**
     * 验证参数签名
     */
    private boolean verifySign(HttpServletRequest httpRequest, HttpServletRequestWrapper requestWrapper,
                               RequestHeader requestHeader) throws IOException {
        SortedMap<String, String> paramMap;
        boolean accept;

        switch (httpRequest.getMethod().toUpperCase()) {
            case "GET":
                paramMap = HttpDataUtil.getUrlParams(requestWrapper);
                accept = SignUtil.verifySign(paramMap, requestHeader);
                break;
            case "POST":
                paramMap = HttpDataUtil.getBodyParams(requestWrapper);
                accept = SignUtil.verifySign(paramMap, requestHeader);
                break;
            default:
                accept = true;
                break;
        }
        return accept;
    }
}
