FROM openjdk:21-jdk-buster

LABEL maintainer="Boolean <hongbin.hsu@bincent.com>"

ADD ./target/*.jar /data/app.jar
COPY config/* /data/config

# set environment
ARG TIME_ZONE="Asia/Shanghai"
ARG BASE_DIR="/data"
ARG JAVA_OPTS="-Xms1g -Xmx1g -Xmn512m -Xss256k -XX:MetaspaceSize=128m -XX:MaxMetaspaceSize=320m -XX:SurvivorRatio=8 -XX:+UseConcMarkSweepGC"

WORKDIR $BASE_DIR

RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list \
	&& apt-get update && apt-get -y upgrade \
    && ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone \
    && mkdir $BASE_DIR/config && mkdir $BASE_DIR/log \
    && rm -rf /var/lib/apt/lists/*

EXPOSE 8080

CMD java -Djava.security.egd=file:/dev/./urandom -jar $BASE_DIR/app.jar  $JAVA_OPTS --appendonly yes