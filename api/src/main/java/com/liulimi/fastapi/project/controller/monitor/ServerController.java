package com.liulimi.fastapi.project.controller.monitor;

import com.liulimi.fastapi.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 服务器监控
 *
 * @author boolean
 */
@RestController
@RequestMapping("/admin/monitor/server")
public class ServerController {

    @GetMapping()
    public ResponseData info() {
        return ResponseData.success();
    }

}
