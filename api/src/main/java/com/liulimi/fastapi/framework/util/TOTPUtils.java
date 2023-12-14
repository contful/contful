package com.liulimi.fastapi.framework.util;

import cn.hutool.core.codec.Base32;
import org.apache.commons.lang3.RandomStringUtils;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.ByteBuffer;
import java.util.UUID;

public class TOTPUtils {
    /**
     * 时间步长，动态口令变化时间周期(单位秒)
     */
    private static final int TIME_STEP = 30;
    /**
     * 动态口令默认长度
     */
    private static final int CODE_DIGITS = 6;

    /**
     * 生成唯一密钥
     * UUID + 4位随机字符生成唯一标识
     *
     * @return String
     */
    public static String generateSecretKey() {
        String uniqueId = UUID.randomUUID() + RandomStringUtils.randomAlphabetic(4);
        return Base32.encode(uniqueId.getBytes()).toLowerCase();
    }

    /**
     * 生成一个基于TOTP标准身份验证器识别的字符串
     * 将该字符串生成二维码可供通用动态密码工具识别，例如:iOS应用(Authy)、微信小程序(二次验证码)
     *
     * @param label
     * @param secret
     * @param issuer
     * @return String
     */
    public static String getQRCodeStr(String label, String secret, String issuer) {
        return String.format("otpauth://totp/%s?secret=%s&issuer=%s", label, secret, issuer);
    }

    /**
     * 生成动态口令
     *
     * @param secret
     * @return
     */
    public static String generateTOTP(String secret) {
        return TOTPUtils.generateTOTP(secret, TOTPUtils.getCurrentInterval(), CODE_DIGITS);
    }

    /**
     * 生成指定位数的动态口令
     *
     * @param secret
     * @param codeDigits
     * @return
     */
    public static String generateTOTP(String secret, int codeDigits) {
        return TOTPUtils.generateTOTP(secret, TOTPUtils.getCurrentInterval(), codeDigits);
    }

    /**
     * 验证动态口令
     *
     * @param secret
     * @param code
     * @return
     */
    public static boolean verify(String secret, String code) {
        return TOTPUtils.verify(secret, code, CODE_DIGITS);
    }

    /**
     * 验证动态口令
     *
     * @param secret
     * @param code
     * @param codeDigits
     * @return
     */
    public static boolean verify(String secret, String code, int codeDigits) {
        long currentInterval = TOTPUtils.getCurrentInterval();
        // 考虑到时间延时，需考虑前一个步长的动态密码是否匹配
        for (int i = 0; i <= 1; i++) {
            String tmpCode = TOTPUtils.generateTOTP(secret, currentInterval - i, codeDigits);
            if (tmpCode.equals(code)) {
                return true;
            }
        }
        return false;
    }

    /**
     * 获取动态口令剩余秒数
     * 所有口令是基于时间戳计算，因此任何动态口令的剩余有效时间都是一致的
     *
     * @return
     */
    public static int getRemainingSeconds() {
        return TIME_STEP - (int) (System.currentTimeMillis() / 1000 % TIME_STEP);
    }

    /**
     * 生成动态口令
     *
     * @param secret
     * @param currentInterval
     * @param codeDigits
     * @return
     */
    public static String generateTOTP(String secret, long currentInterval, int codeDigits) {
        if (codeDigits < 1 || codeDigits > 18) {
            throw new UnsupportedOperationException("不支持" + codeDigits + "位数的动态口令");
        }
        byte[] content = ByteBuffer.allocate(8).putLong(currentInterval).array();
        byte[] hash = TOTPUtils.hmacsha(content, secret);
        // 获取hash最后一个字节的低4位，作为选择结果的开始下标偏移
        int offset = 0;
        if (hash != null) {
            offset = hash[hash.length - 1] & 0xf;
        }
        // 获取4个字节组成一个整数，其中第一个字节最高位为符号位，不获取，使用0x7f
        int binary = 0;
        if (hash != null) {
            binary = ((hash[offset] & 0x7f) << 24) |
                    ((hash[offset + 1] & 0xff) << 16) |
                    ((hash[offset + 2] & 0xff) << 8) |
                    (hash[offset + 3] & 0xff);
        }
        // 如果所需位数为6，则该值为1000000
        long digitsPower = Long.parseLong(TOTPUtils.rightPadding("1", codeDigits + 1));
        // 获取当前数值后的指定位数
        long code = binary % digitsPower;
        // 将数字转成字符串，不够指定位前面补0
        return TOTPUtils.leftPadding(Long.toString(code), codeDigits);
    }

    /**
     * 获取当前时间戳
     *
     * @return
     */
    private static long getCurrentInterval() {
        return System.currentTimeMillis() / 1000 / TIME_STEP;
    }

    /**
     * 向左补足0
     *
     * @param value
     * @param length
     * @return
     */
    private static String leftPadding(String value, int length) {
        StringBuilder valueBuilder = new StringBuilder(value);
        while (valueBuilder.length() < length) {
            valueBuilder.insert(0, "0");
        }
        value = valueBuilder.toString();
        return value;
    }

    /**
     * 向右补足0
     *
     * @param value
     * @param length
     * @return
     */
    private static String rightPadding(String value, int length) {
        StringBuilder valueBuilder = new StringBuilder(value);
        while (valueBuilder.length() < length) {
            valueBuilder.append("0");
        }
        value = valueBuilder.toString();
        return value;
    }

    /**
     * 使用HmacSHA1加密
     *
     * @param content
     * @param key
     * @return
     */
    private static byte[] hmacsha(byte[] content, String key) {
        try {
            byte[] byteKey = Base32.decode(key);
            Mac hmac = Mac.getInstance("HmacSHA1");
            SecretKeySpec keySpec = new SecretKeySpec(byteKey, "HmacSHA1");
            hmac.init(keySpec);
            return hmac.doFinal(content);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
}
