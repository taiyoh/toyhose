# AWS SDK for Go v2 Migration Guide

This document records the process and key learnings from migrating the `toyhose` project from AWS SDK for Go v1 to v2.

## 1. Motivation

The primary motivation for this migration was that AWS SDK for Go v1 is no longer actively maintained. Migrating to v2 is essential for long-term security, performance, and access to new AWS features.

## 2. Key Migration Challenges and Solutions

The migration was not a simple one-to-one replacement and involved several key challenges.

### Challenge 1: Signature V4 Verification

- **Problem**: The project's API request verification (`verifier_v4.go`) relied on the v1 `signer/v4` package to perform server-side signature recalculation. Initial investigation suggested that v2 might not expose a similar public API, which was the biggest potential blocker.
- **Solution**: Further research confirmed that v2 **does** provide a public signing package at `github.com/aws/aws-sdk-go-v2/aws/signer/v4`. The verification logic was successfully migrated by replacing the v1 signer with the v2 signer. A key part of the solution was to mimic v1's behavior by carefully filtering HTTP headers included in the signature calculation.

### Challenge 2: Client Initialization

- **Problem**: v1 used a `session` object (`session.NewSession()`) to create service clients. This concept does not exist in v2.
- **Solution**: The entire client and configuration initialization process was refactored to use the v2 paradigm. This involved:
  - Using `config.LoadDefaultConfig` to load configuration from the environment.
  - Providing credentials via `credentials.NewStaticCredentialsProvider`.
  - Creating service clients directly from the `aws.Config` object (e.g., `s3.NewFromConfig(cfg)`).

### Challenge 3: Error Handling

- **Problem**: v1 relied on type assertions to the `awserr.Error` interface to inspect error codes.
- **Solution**: Error handling was updated to the standard Go 1.13+ approach using `errors.As`. This allows for more robust, type-safe error checking (e.g., `errors.As(err, &resourceNotFoundException)`).

### Challenge 4: API Call Signatures

- **Problem**: Most v1 API methods accepted `*some.Input` as jejich argument. v2 methods require `context.Context` as the first argument.
- **Solution**: All API calls throughout the codebase were updated to pass a `context.Context` as the first parameter.

### Challenge 5: Timestamp Deserialization

- **Problem**: After migrating the code, tests failed with a `deserialization failed` error. Specifically, the test client (using SDK v2) expected timestamp fields to be a JSON Number (Unix timestamp), but the `toyhose` server was encoding them as RFC3339 strings, which is the default for Go's `encoding/json`.
- **Solution**: The `DescribeDeliveryStream` method in `delivery_stream_service.go` was modified to return a custom struct for JSON marshaling. In this struct, the `CreateTimestamp` field is explicitly converted from `time.Time` to a Unix timestamp (`int64`) before being encoded to JSON, resolving the format mismatch.
