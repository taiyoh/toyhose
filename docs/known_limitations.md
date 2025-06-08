# Known Limitations

This document lists the known functional and operational limitations of the current version of `toyhose`.

## 1. Error Handling and Data Loss

- **`ErrorOutputPrefix` is Not Implemented**: The `ErrorOutputPrefix` feature, which is part of the Firehose API, is not functionally implemented. There is currently no mechanism to separate and reroute records that fail processing.
- **Potential Data Loss on S3 Failure**: If `toyhose` fails to deliver a batch of records to S3 after 30 retry attempts, the entire batch is discarded and logged as an error. This results in data loss for that batch.

## 2. API and Feature Coverage

- **Limited Destination Support**: The only supported destination is Amazon S3 (`S3DestinationConfiguration`). Other destinations like Elasticsearch, Redshift, and Splunk are not supported.
- **No Data Transformation**: `toyhose` does not support data transformation with AWS Lambda, which is a feature of the real Firehose service. All records are passed through to the destination as-is.
- **Unsupported API Operations**: Several API operations related to tagging, encryption, and destination updates are not implemented. Please refer to the [Roadmap](./roadmap.md) for a complete list.

## 3. Performance and Scalability

- **Single-Node Architecture**: `toyhose` runs as a single process and is not designed for horizontal scalability or high-availability clusters. It is intended for local development and testing, not for large-scale production workloads.
- **In-Memory Buffering**: Record buffering is handled entirely in memory. If the `toyhose` process crashes, any data currently held in the buffer will be lost.
