# Library Choices

This document outlines the rationale behind the selection of key external libraries in the `toyhose` project.

### `rs/zerolog`

- **Purpose**: Structured, high-performance logging.
- **Rationale**:
  - **Performance**: `zerolog` is known for its low allocation overhead, making it one of the fastest logging libraries in the Go ecosystem. This is beneficial in a high-throughput service like `toyhose`.
  - **Structured Logging**: It produces JSON-formatted logs by default. Structured logs are machine-parsable, which greatly simplifies log analysis, searching, and integration with modern log management systems (e.g., ELK stack, Datadog). This is a significant advantage over plain-text logs for debugging and monitoring.

### `caarlos0/env`

- **Purpose**: Parsing environment variables into Go structs.
- **Rationale**:
  - **Simplicity and Declarativeness**: This library allows developers to define configuration schemas as Go structs with `env` tags. It cleanly separates configuration definition from the application logic and removes the need for boilerplate code to manually parse each environment variable.
  - **Type Safety**: It handles the conversion of string-based environment variables into various Go types (int, bool, etc.), including handling default values and required fields.

### `google/uuid`

- **Purpose**: Generating universally unique identifiers (UUIDs).
- **Rationale**:
  - **Standard and Robust**: This is a widely used and well-tested library from Google for creating standard UUIDs.
  - **Uniqueness**: In `toyhose`, it's used to generate unique filenames for S3 objects, which is critical to prevent data from being overwritten when multiple batches are delivered in the same time window.

### `vjeantet/jodaTime`

- **Purpose**: Formatting time based on Joda-Time/strftime patterns.
- **Rationale**:
  - **Developer Experience**: The standard Go `time.Format` uses a unique reference-date-based layout (`2006-01-02`), which can be unintuitive for developers accustomed to `YYYY-MM-DD` style patterns common in other languages.
  - **Feature Requirement**: This library was chosen to implement the `!{timestamp:YYYY-MM-dd}` formatting feature for S3 prefixes, providing a familiar and flexible way for users to define their S3 object key structure.
