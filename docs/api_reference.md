# API Reference

This document outlines the supported AWS Kinesis Firehose API endpoints in `toyhose`.

| API Endpoint | Supported | Notes |
|---|---|---|
| `CreateDeliveryStream` | 🙆‍♀️ Yes | Supports `KinesisStreamSourceConfiguration` and `S3DestinationConfiguration`. Other destination types are not implemented. |
| `DeleteDeliveryStream` | 🙆‍♀️ Yes | |
| `DescribeDeliveryStream` | 🙆‍♀️ Yes | |
| `ListDeliveryStreams` | 🙆‍♀️ Yes | |
| `ListTagsForDeliveryStream` | 🙊 No | |
| `PutRecord` | 🙆‍♀️ Yes | |
| `PutRecordBatch` | 🙆‍♀️ Yes | |
| `StartDeliveryStreamEncryption` | 🙊 No | |
| `StopDeliveryStreamEncryption` | 🙊 No | |
| `TagDeliveryStream` | 🙊 No | |
| `UntagDeliveryStream` | 🙊 No | |
| `UpdateDestination` | 🙊 No | |

## `CreateDeliveryStream` Details

### Supported Configurations

- **`DeliveryStreamType`**: `KinesisStreamAsSource` is the only supported type. Direct PUT is implicitly supported.
- **`KinesisStreamSourceConfiguration`**:
  - `KinesisStreamARN`: The ARN of the source Kinesis Data Stream.
  - `RoleARN`: Required by the AWS API, but not actually used by `toyhose`.
- **`S3DestinationConfiguration`**:
  - `BucketARN`: The ARN of the destination S3 bucket.
  - `RoleARN`: Required by the AWS API, but not actually used by `toyhose`.
  - `BufferingHints`:
    - `IntervalInSeconds`: Time to buffer data before delivery to S3.
    - `SizeInMBs`: Size of data to buffer before delivery.
  - `CompressionFormat`: `GZIP`, `UNCOMPRESSED`.
  - `Prefix`: A prefix for S3 object keys.
  - `ErrorOutputPrefix`: A prefix for S3 object keys for records that failed processing.

### Unsupported Configurations

- `ElasticsearchDestinationConfiguration`
- `ExtendedS3DestinationConfiguration` (Note: This is on the roadmap 👷)
- `RedshiftDestinationConfiguration`
- `SplunkDestinationConfiguration`
