# Configuration

`toyhose` is configured using environment variables. The following variables are available to customize its behavior.

## 1. AWS Credentials and Region

These variables are used to configure the AWS client that `toyhose` uses to interact with other services like Kinesis and S3.

- `AWS_ACCESS_KEY_ID` ( **required** ): The AWS access key ID.
- `AWS_SECRET_ACCESS_KEY` ( **required** ): The AWS secret access key.
- `AWS_REGION` (optional, default: `us-east-1`): The AWS region.

## 2. Server Configuration

- `PORT` (optional, default: `4573`): The port on which the `toyhose` server will listen for API requests. The default is inspired by LocalStack's Firehose port.

## 3. S3 Destination Configuration

These variables control the behavior of the S3 destination, including buffering and endpoint overrides.

- `S3_ENDPOINT_URL` (optional): The endpoint URL for the S3 service. Use this to target a local S3-compatible service like MinIO or LocalStack (e.g., `http://localhost:4566`). If not set, it defaults to the standard AWS S3 endpoint for the specified region.
- `S3_DISABLE_BUFFERING` (optional, default: `false`): If set to `true`, buffering is disabled, and records are delivered to S3 immediately. This is useful for testing but not recommended for production-like scenarios.
- `S3_BUFFERING_HINTS_SIZE_IN_MBS` (optional): Overrides the `SizeInMBs` buffering hint set in `CreateDeliveryStream`.
- `S3_BUFFERING_HINTS_INTERVAL_IN_SECONDS` (optional): Overrides the `IntervalInSeconds` buffering hint set in `CreateDeliveryStream`.

## 4. Kinesis Source Configuration

- `KINESIS_STREAM_ENDPOINT_URL` (optional): The endpoint URL for the Kinesis Data Streams service. Use this to target a local Kinesis-compatible service like LocalStack (e.g., `http://localhost:4566`). If not set, it defaults to the standard AWS Kinesis endpoint.

## Example `docker-compose.yml`

```yaml
version: "3.8"
services:
  toyhose:
    image: your-toyhose-image # Replace with your actual image
    ports:
      - "4573:4573"
    environment:
      - AWS_ACCESS_KEY_ID=test
      - AWS_SECRET_ACCESS_KEY=test
      - AWS_REGION=us-east-1
      - S3_ENDPOINT_URL=http://localstack:4566
      - KINESIS_STREAM_ENDPOINT_URL=http://localstack:4566
  
  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      - SERVICES=s3,kinesis
