FROM apache/flink:1.19.1-scala_2.12-java17

ARG FLINK_KAFKA_CONNECTOR_VERSION=3.2.0-1.19
ARG ICEBERG_VERSION=1.6.1
ARG HADOOP_VERSION=3.3.6
ARG AWS_SDK_BUNDLE_VERSION=1.12.720

RUN curl -fLo /opt/flink/lib/flink-sql-connector-kafka.jar \
      https://repo1.maven.org/maven2/org/apache/flink/flink-sql-connector-kafka/${FLINK_KAFKA_CONNECTOR_VERSION}/flink-sql-connector-kafka-${FLINK_KAFKA_CONNECTOR_VERSION}.jar \
    && curl -fLo /opt/flink/lib/iceberg-flink-runtime.jar \
      https://repo1.maven.org/maven2/org/apache/iceberg/iceberg-flink-runtime-1.19/${ICEBERG_VERSION}/iceberg-flink-runtime-1.19-${ICEBERG_VERSION}.jar \
    && curl -fLo /opt/flink/lib/hadoop-aws.jar \
      https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/${HADOOP_VERSION}/hadoop-aws-${HADOOP_VERSION}.jar \
    && curl -fLo /opt/flink/lib/aws-java-sdk-bundle.jar \
      https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/${AWS_SDK_BUNDLE_VERSION}/aws-java-sdk-bundle-${AWS_SDK_BUNDLE_VERSION}.jar
