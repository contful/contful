package com.contful.framework.annotation;

import java.lang.annotation.*;

/**
 * 重放验证
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
@Documented
public @interface ReplayCheck {
}
