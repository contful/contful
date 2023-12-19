package com.contful.admin.controller.account;

import com.contful.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/admin/account/profile")
public class ProfileController {

    /**
     * 个人信息
     * @return
     */
    @GetMapping("/")
    public ResponseData get() {
        return ResponseData.success();
    }

    /**
     * 修改个人信息
     * @return
     */
    @PutMapping("/")
    public ResponseData update() {
        return ResponseData.success();
    }
}
