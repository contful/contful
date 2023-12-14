package com.liulimi.fastapi.framework.annotation;

import java.lang.annotation.*;

/**
 * 忽略签名验证
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
@Documented
public @interface SignIgnore {
}
