# Architecture

This document describes the high-level architecture of `toyhose`.

## 1. Component Overview

`toyhose` consists of several key components that work together to emulate the AWS Kinesis Firehose service.

- **HTTP Dispatcher (`dispatcher.go`)**: The entry point for all incoming API requests. It parses the request target (e.g., `Firehose_20150804.CreateDeliveryStream`) and routes the request payload to the appropriate service.
- **Delivery Stream Service (`delivery_stream_service.go`)**: Contains the business logic for the Firehose API operations (`CreateDeliveryStream`, `PutRecord`, etc.). It manages the lifecycle of delivery streams.
- **Delivery Stream (`delivery_stream.go`)**: Represents a single Firehose delivery stream. It holds the stream's configuration and manages the underlying data source and destination.
- **Kinesis Consumer (`kinesis_consumer.go`)**: When a delivery stream is configured with a Kinesis Data Stream as its source, this component is responsible for consuming records from that stream.
- **S3 Destination (`s3_destination.go`)**: Manages the buffering of records and their eventual delivery to the configured S3 bucket. It handles buffering based on time and size, data compression, and writing objects to S3.

## 2. Data Flow

Here are the two primary data flows in `toyhose`:

### Data Flow 1: Direct PUT (`PutRecord`/`PutRecordBatch`)

1.  An HTTP request for `PutRecord` or `PutRecordBatch` hits the **Dispatcher**.
2.  The **Dispatcher** forwards the request to the `Put` or `PutBatch` method in the **Delivery Stream Service**.
3.  The service finds the target **Delivery Stream** and passes the records to it.
4.  The records are sent to the `recordCh` channel, which is monitored by the **S3 Destination**.
5.  The **S3 Destination** component buffers the records.
6.  When the buffer is full (either by size or time), the **S3 Destination** writes the buffered data as a single object to the configured S3 bucket.

### Data Flow 2: Kinesis Stream as Source

1.  A delivery stream is created with `KinesisStreamSourceConfiguration`.
2.  The **Delivery Stream** component instantiates a **Kinesis Consumer**.
3.  The **Kinesis Consumer** starts a long-running process to poll the specified Kinesis Data Stream for new records.
4.  As records are fetched from Kinesis, they are sent to the `recordCh` channel.
5.  From this point, the process is the same as the Direct PUT flow (steps 5 and 6). The **S3 Destination** buffers and writes the data to S3.

## 3. Diagram

```mermaid
graph TD
    subgraph "Client"
        A[API Client]
    end

    subgraph "toyhose Service"
        B[Dispatcher]
        C[DeliveryStreamService]
        D[DeliveryStream]
        E[S3Destination]
        F[KinesisConsumer]
    end

    subgraph "AWS Services (or local equivalent)"
        G[Kinesis Data Stream]
        H[S3 Bucket]
    end

    A -- HTTP API Request --> B
    B -- Routes to --> C

    C -- Manages --> D

    A -- PutRecord(Batch) --> B
    B -- to --> C
    C -- sends records to --> D
    D -- recordCh --> E

    F -- Polls for records --> G
    F -- sends records to --> D

    D -- Has a --> F
    D -- Has a --> E

    E -- Writes object to --> H
