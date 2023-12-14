package com.liulimi.fastapi.framework.indicator;

import com.liulimi.fastapi.framework.condition.RedisHealthCondition;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.context.annotation.Conditional;
import org.springframework.stereotype.Component;

@Component
@Conditional(RedisHealthCondition.class)
public class RedisHealthIndicator implements HealthIndicator {
    @Override
    public Health health() {
        return Health.up().build();
    }
}
