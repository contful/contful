package com.contful.framework.annotation;

import java.lang.annotation.*;

/**
 * 将方法标记为非线程安全的注解.
 *
 * @author Boolean
 */
@Documented
@Target({ElementType.TYPE, ElementType.METHOD})
@Retention(RetentionPolicy.SOURCE)
public @interface NotThreadSafe {

}