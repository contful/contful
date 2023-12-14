package com.liulimi.fastapi.framework.config;

import com.liulimi.fastapi.framework.interceptor.AuthorizationInterceptor;
import com.liulimi.fastapi.framework.interceptor.SignInterceptor;
import jakarta.annotation.Resource;
import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.http.MediaType;
import org.springframework.web.servlet.config.annotation.ContentNegotiationConfigurer;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;


/**
 * 注册拦截器的配置类
 *
 * @author boolean 2022-08-17
 */
@AutoConfiguration
public class InterceptorConfig implements WebMvcConfigurer {
    @Resource
    private SignInterceptor signInterceptor;

    @Resource
    private AuthorizationInterceptor authorizationInterceptor;

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(signInterceptor)
                .addPathPatterns("/**"); // 拦截所有路径

        registry.addInterceptor(authorizationInterceptor) // 注册自定义拦截器
                .addPathPatterns("/**") // 拦截所有路径
                .excludePathPatterns("/favicon.ico")
                .excludePathPatterns("/test");
    }

    @Override
    public void addCorsMappings(CorsRegistry registry) {
        //对那些请求路径进行跨域处理
        registry.addMapping("/**")
                // 允许访问的客户端域名
                .allowedOriginPatterns("*")
                // 允许访问的方法名
                .allowedMethods("GET", "POST", "DELETE", "PUT", "PATCH")
                // 允许服务端访问的客户端请求头
                .allowedHeaders("*")
                .allowCredentials(true)
                .maxAge(3600);
    }

    /**
     * 自定义内容协商策略
     */
    @Override
    public void configureContentNegotiation(ContentNegotiationConfigurer configurer) {
        configurer.favorParameter(true)        //是否支持请求参数的方式
                .ignoreAcceptHeader(true)   // 不检查Accept请求头
                .parameterName("format")    //请求参数名
                .defaultContentType(MediaType.APPLICATION_JSON) //默认返回格式
                .mediaType("json", MediaType.APPLICATION_JSON)
                .mediaType("xml", MediaType.APPLICATION_ATOM_XML);
    }
}