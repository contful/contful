package com.contful.admin.controller.account;

import com.contful.framework.annotation.AuthIgnore;
import com.contful.framework.annotation.SignIgnore;
import com.contful.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 账号服务
 */
@RestController
@RequestMapping("/admin/account")
public class AccountController {

    /**
     * 用户注册
     *
     * @return
     */
    @SignIgnore
    @AuthIgnore
    @PostMapping("/signup")
    public ResponseData signup() {
        return ResponseData.success();
    }

    /**
     * 账号登陆
     *
     * @return
     */
    @SignIgnore
    @AuthIgnore
    @PostMapping("/signin")
    public ResponseData signin() {
        return ResponseData.success();
    }

    /**
     * 用户退出
     *
     * @return
     */
    @PostMapping("/signout")
    public ResponseData signout() {
        return ResponseData.success();
    }
}
