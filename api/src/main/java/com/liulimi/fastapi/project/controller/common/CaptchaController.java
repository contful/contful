package com.liulimi.fastapi.project.controller.common;

import com.liulimi.fastapi.framework.annotation.AuthIgnore;
import com.liulimi.fastapi.framework.annotation.SignIgnore;
import com.liulimi.fastapi.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 验证码操作处理
 */
@RestController
@RequestMapping("/admin/common/captcha")
public class CaptchaController {
    /**
     * 获取图片验证码
     * @return
     */
    @SignIgnore
    @AuthIgnore
    @PostMapping("/image")
    public ResponseData image() {
        return ResponseData.success();
    }
}
