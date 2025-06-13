# Error Handling

This document describes the current error handling capabilities in `toyhose`.

## 1. S3 Destination Error Handling

The primary error handling mechanism within the S3 destination involves retrying failed `PutObject` operations.

### Current Behavior

- **S3 PutObject Retries**: When `toyhose` attempts to write a batch of records to S3, if the `PutObject` API call fails, it will retry the operation up to 30 times with a 100ms delay between each attempt.
- **Data Loss on Failure**: If all 30 retries fail, the batch of records is discarded, and an error is logged. The data within that batch is lost.
- **No Per-Record Processing**: The system does not currently inspect or process individual records within a batch. The entire batch is treated as a single unit.

### Unimplemented Features (`ErrorOutputPrefix`)

The `CreateDeliveryStream` API accepts an `ErrorOutputPrefix` parameter in the `S3DestinationConfiguration`. The intention of this feature in AWS Kinesis Firehose is to deliver records that failed processing to a separate S3 location.

However, in the current version of `toyhose`, **this feature is not implemented**.

- The `ErrorOutputPrefix` value is stored but is **never used**.
- There is no logic to identify, separate, or reroute "failed" records because no per-record processing (like schema validation or data transformation) is performed.

## 2. API-Level Errors

Standard API-level errors are returned for invalid requests, such as:
- Requesting a non-existent delivery stream.
- Providing invalid parameters in a `CreateDeliveryStream` call.

These errors are generally aligned with the error responses of the actual AWS Kinesis Firehose service.
