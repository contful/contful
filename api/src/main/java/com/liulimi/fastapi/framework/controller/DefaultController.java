package com.liulimi.fastapi.framework.controller;

import com.liulimi.fastapi.framework.annotation.AuthIgnore;
import com.liulimi.fastapi.framework.annotation.SignIgnore;
import com.liulimi.fastapi.framework.domain.ResponseData;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class DefaultController {

    @SignIgnore
    @AuthIgnore
    @GetMapping(value = "/timestamp")
    public ResponseData timestamp() {
        return ResponseData.success();
    }
}
