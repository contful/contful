package com.contful.framework.generator;

import com.baomidou.mybatisplus.generator.AutoGenerator;
import com.baomidou.mybatisplus.generator.config.*;
import com.baomidou.mybatisplus.generator.config.rules.NamingStrategy;

/**
 * 使用MyBatisPlus生成器生成数据库CRUD代码
 */
public class CodeGeneratorTest {

    public static void main(String[] args) {
        // 配置 DataSourceConfig
        String url = "jdbc:mysql://127.0.0.1:3306/test";
        String username = "root";
        String password = "";
        DataSourceConfig dataSourceConfig = new DataSourceConfig.Builder(url, username, password).build();

        // 配置 GlobalConfig
        String projectPath = System.getProperty("user.dir");
        System.out.println(projectPath);
        GlobalConfig globalConfig = new GlobalConfig.Builder()
                .outputDir(projectPath + "/src/main/java") //设置输出路径
                .author("FastAPI")
                .build();

        // 包名配置
        PackageConfig packageConfig = new PackageConfig.Builder()
                .moduleName("test")
                .parent("com.liulimi.fastapi")
                .entity("entity.bean")
                .build();

        // 策略配置
        StrategyConfig strategyConfig = new StrategyConfig.Builder()
                .enableCapitalMode()
                .controllerBuilder()
                .enableRestStyle()  // 配置 rest 风格的控制器（@RestController）
                .enableHyphenStyle()    // 配置驼峰转连字符
                .entityBuilder()
                .enableLombok() // 使用lombok
                .naming(NamingStrategy.underline_to_camel)
                .columnNaming(NamingStrategy.underline_to_camel)
                //.enableRemoveIsPrefix() // 设置布尔类型字段是否去掉is前缀
                .build();


        // 自定义代码模板
        TemplateConfig templateConfig = new TemplateConfig.Builder()
                //.entity("templates/entity/bean.vm")
                .build();

        AutoGenerator autoGenerator = new AutoGenerator(dataSourceConfig);
        autoGenerator.global(globalConfig)
                .strategy(strategyConfig)
                .packageInfo(packageConfig)
                .template(templateConfig)
                .execute();
    }
}
