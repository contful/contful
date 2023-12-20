package com.contful.framework.condition;

import com.contful.framework.util.StringUtil;
import org.springframework.context.annotation.Condition;
import org.springframework.context.annotation.ConditionContext;
import org.springframework.core.type.AnnotatedTypeMetadata;

public class RedisHealthCondition implements Condition {
    /**
     * 判断配置文件中是否存在Redis配置
     *
     * @param context  条件上下文
     * @param metadata 元数据
     * @return boolean
     */
    @Override
    public boolean matches(ConditionContext context, AnnotatedTypeMetadata metadata) {
        String redisProperty = context.getEnvironment().getProperty("spring.data.redis.host");
        return StringUtil.isBlank(redisProperty);
    }
}
