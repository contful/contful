package com.contful.admin.controller.common;

import com.contful.framework.annotation.AuthIgnore;
import com.contful.framework.annotation.SignIgnore;
import com.contful.framework.domain.ResponseData;
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
