# Project Overview

## 1. Purpose

`toyhose` is a lightweight, open-source emulation of AWS Kinesis Firehose. It is designed to provide a local development and testing environment for applications that interact with Kinesis Firehose.

The primary goal is to enable developers to test their data ingestion pipelines without needing to connect to a live AWS environment, thereby reducing costs and improving development speed.

## 2. Key Features

- **Firehose API Emulation**: Implements a subset of the Kinesis Firehose API, allowing applications to create, manage, and send data to delivery streams.
- **Kinesis Stream Source**: Can be configured to pull data from an AWS Kinesis Data Stream, simulating a common data pipeline pattern.
- **S3 Destination**: Buffers incoming records and delivers them to a local or AWS S3 bucket, mimicking the behavior of a real Firehose delivery stream.
- **Configurable**: Supports configuration for buffering hints (size and interval) and S3 object key prefixes.

## 3. Use Cases

- **Local Development**: Develop and debug applications that produce data for Firehose without incurring AWS costs.
- **Integration Testing**: Run automated integration tests for data pipelines in a CI/CD environment.
- **Offline Work**: Continue development even when an internet connection to AWS is unavailable.
