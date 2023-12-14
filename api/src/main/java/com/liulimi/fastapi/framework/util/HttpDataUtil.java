package com.liulimi.fastapi.framework.util;

import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import jakarta.servlet.http.HttpServletRequest;
import lombok.experimental.UtilityClass;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.springframework.util.StreamUtils;

import java.io.IOException;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.util.SortedMap;
import java.util.TreeMap;

@Slf4j
@UtilityClass
public class HttpDataUtil {
    /**
     * post请求处理：获取 Body 参数，转换为SortedMap
     */
    public SortedMap<String, String> getBodyParams(final HttpServletRequest request) throws IOException {
        byte[] requestBody = StreamUtils.copyToByteArray(request.getInputStream());
        JSONObject jsonObject = JSONUtil.parseObj(requestBody);
        SortedMap<String, String> bodyParams = new TreeMap<>();

        for (String key : jsonObject.keySet()) {
            String value = (String) jsonObject.get(key);
            bodyParams.put(key, value);
        }
        return bodyParams;
    }

    /**
     * get请求处理：将URL请求参数转换成SortedMap
     */
    public static SortedMap<String, String> getUrlParams(HttpServletRequest request) {
        String param = "";
        SortedMap<String, String> result = new TreeMap<>();

        if (StringUtils.isEmpty(request.getQueryString())) {
            return result;
        }

        param = URLDecoder.decode(request.getQueryString(), StandardCharsets.UTF_8);

        String[] params = param.split("&");
        for (String s : params) {
            String[] array = s.split("=");
            result.put(array[0], array[1]);
        }

        return result;
    }
}
