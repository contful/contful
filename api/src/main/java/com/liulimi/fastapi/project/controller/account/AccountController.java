package com.liulimi.fastapi.project.controller.account;

import com.liulimi.fastapi.framework.annotation.AuthIgnore;
import com.liulimi.fastapi.framework.annotation.SignIgnore;
import com.liulimi.fastapi.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.*;

/**
 * 账号服务
 */
@RestController
@RequestMapping("/admin/account")
public class AccountController {

    /**
     * 用户注册
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
     * @return
     */
    @PostMapping("/signout")
    public ResponseData signout() {
        return ResponseData.success();
    }
}
