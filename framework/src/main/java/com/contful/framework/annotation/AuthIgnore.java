package com.contful.framework.annotation;

import java.lang.annotation.*;

/**
 * 忽略Token验证
 *
 * @author Boolean 2021-04-05
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
@Documented
public @interface AuthIgnore {
}