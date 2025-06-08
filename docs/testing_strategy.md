# Testing Strategy

The `toyhose` project employs a two-pronged testing strategy to ensure code quality, correctness, and reliability: **Unit Tests** and **Integration Tests**.

## 1. Unit Tests

- **Purpose**: To test individual components (functions or methods) in isolation. These tests focus on the internal logic of a component without involving external dependencies like network services.
- **Location**: Found in `*_test.go` files, such as `s3_prefix_test.go`.
- **Characteristics**:
  - Fast execution speed.
  - No external dependencies required.
  - Focus on testing specific logic, edge cases, and input validation (e.g., testing the S3 key prefix generation logic in `keyPrefix`).

## 2. Integration Tests

- **Purpose**: To verify that different components of `toyhose` work together correctly, including its interactions with external services that it emulates (S3 and Kinesis).
- **Location**: Found in `*_test.go` files that test the interaction between components, such as `s3_destination_test.go` and `kinesis_consumer_test.go`.
- **Characteristics**:
  - Slower execution compared to unit tests.
  - Require a running environment with dependencies.
  - Focus on testing end-to-end data flows.

### Test Environment Setup

Integration tests rely on a Docker-based environment defined in `docker-compose.yml`. This environment provides local, lightweight emulations of AWS services:

- **S3**: `minio/minio` is used as an S3-compatible object storage.
- **Kinesis Data Streams**: `instructure/kinesalite` is used as a Kinesis-compatible data stream service.

### Test Execution Flow

The integration tests follow a clear setup and teardown pattern, managed by helper functions in `aws_setup_test.go`:

1.  **Test Start**: Before a test runs, a new, temporary S3 bucket or Kinesis stream is created on the MinIO or Kinesalite container using functions like `setupS3` and `setupKinesisStream`.
2.  **Test Logic**: The test executes its logic against this temporary resource (e.g., delivering a batch of records to the S3 bucket).
3.  **Test End (Cleanup)**: The `t.Cleanup` function ensures that the temporary S3 bucket or Kinesis stream is automatically deleted after the test completes, regardless of whether it passed or failed.

This approach guarantees that each integration test runs in a clean, isolated environment, preventing interference between tests and ensuring reliable, repeatable results.
