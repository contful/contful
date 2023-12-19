package com.contful.framework.controller;

import com.contful.framework.annotation.AuthIgnore;
import com.contful.framework.annotation.SignIgnore;
import com.contful.framework.domain.ResponseData;
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
