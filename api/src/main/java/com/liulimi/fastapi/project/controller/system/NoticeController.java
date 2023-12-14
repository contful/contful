package com.liulimi.fastapi.project.controller.system;

import com.liulimi.fastapi.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/admin/system/notice")
public class NoticeController {

    /**
     * 通知列表
     * @return
     */
    @GetMapping("/items")
    public ResponseData items() {
        return ResponseData.success();
    }

    /**
     * 通知详情
     * @return
     */
    @GetMapping("/detail")
    public ResponseData detail() {
        return ResponseData.success();
    }
}
