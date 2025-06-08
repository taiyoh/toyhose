# Project Roadmap

This document outlines the current implementation status and future development plans for `toyhose`, based on the AWS Kinesis Firehose API.

| Status Icon | Meaning |
|---|---|
| 🙆‍♀️ | Implemented |
| 👷 | Under Construction (Planned) |
| 🙊 | Not Implemented (Not Planned) |

## Implemented Features (🙆‍♀️)

The following API operations and configurations are currently supported:

- **Core API**:
  - `CreateDeliveryStream`
  - `DeleteDeliveryStream`
  - `DescribeDeliveryStream`
  - `ListDeliveryStreams`
  - `PutRecord`
  - `PutRecordBatch`
- **`CreateDeliveryStream` Configurations**:
  - `KinesisStreamSourceConfiguration` (Kinesis Data Stream as a source)
  - `S3DestinationConfiguration` (S3 as a destination)

## Planned Features (👷)

The following features are on the development roadmap:

- **`ExtendedS3DestinationConfiguration`**: This will add support for more advanced S3 destination settings, likely including custom S3 object key formats and more sophisticated error output configurations.

## Not Planned (🙊)

The following features are currently out of scope for the `toyhose` project. This is generally because they are less critical for the primary use case of local development and testing of common data pipelines.

- **Alternative Destinations**:
  - `ElasticsearchDestinationConfiguration`
  - `RedshiftDestinationConfiguration`
  - `SplunkDestinationConfiguration`
- **Tagging and Encryption**:
  - `ListTagsForDeliveryStream`
  - `TagDeliveryStream`
  - `UntagDeliveryStream`
  - `StartDeliveryStreamEncryption`
  - `StopDeliveryStreamEncryption`
- **Updates**:
  - `UpdateDestination`
