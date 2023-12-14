package com.liulimi.fastapi.project.controller.account;

import com.liulimi.fastapi.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/admin/account/password")
public class PasswordController
{
    /**
     * 重置密码
     * @return
     */
    @PutMapping("/reset")
    public ResponseData reset() {
        return ResponseData.success();
    }
}
