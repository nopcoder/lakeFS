---
layout: default
title: Using Python with lakeFS - Overview
description: Guide to using Python for interacting with lakeFS, covering official SDKs and other integration methods.
parent: Using lakeFS with...
nav_order: 30 # Or adjust as needed
has_children: true
---

# Using Python with lakeFS

Python is a first-class citizen for interacting with lakeFS, offering several ways to integrate lakeFS into your Python applications and data workflows. Whether you prefer a high-level idiomatic Python experience or direct access to the lakeFS API, there's a client that fits your needs.

This guide is broken down into the following sections:

## Python Integration Options

lakeFS offers a few ways to work with Python:

1.  **High-Level `lakefs` SDK (Recommended):** Provides a user-friendly, idiomatic Python interface for common lakeFS operations like managing repositories, branches, objects, and performing version control actions. This is generally the recommended client for most use cases. ([Full Guide](./python/hl_sdk_overview.md))
2.  **Generated `lakefs-sdk`:** A Python client library that is auto-generated directly from the lakeFS OpenAPI specification. It provides comprehensive coverage of the lakeFS API. ([Full Guide](./python/generated_sdk.md))
3.  **`lakefs-spec`:** An fsspec-compatible file system implementation for lakeFS. Ideal for integrating with Python data libraries like Pandas and Dask. ([Full Guide](./python/lakefs_spec.md))
4.  **`boto3` (via S3 Gateway):** Use `boto3` to interact with objects in lakeFS via its S3-compatible gateway. ([Full Guide](./python/boto3_usage.md))

**Note on Legacy Clients:**
Previous versions of this documentation and some older community examples might refer to using libraries like `bravado` to dynamically generate a client from a `swagger.json` file. While this approach works, the officially supported and recommended clients are the High-Level `lakefs` SDK and the generated `lakefs-sdk` mentioned above, as they offer better support, performance, and an improved developer experience.

## Detailed Guides

Explore the specifics of each Python integration method:

### High-Level `lakefs` SDK

*   [Overview, Installation, and Initialization](./python/hl_sdk_overview.md)
*   [Working with Repositories](./python/hl_sdk_repositories.md)
*   [Working with Branches](./python/hl_sdk_branches.md)
*   [Working with Objects (Files I/O)](./python/hl_sdk_objects.md)
*   [Committing Changes](./python/hl_sdk_commits.md)
*   [Merging and Tagging](./python/hl_sdk_merging_tags.md)
*   [Transactions and Importing Data](./python/hl_sdk_transactions_import.md)
*   [Error Handling and API Reference](./python/hl_sdk_error_handling.md)

### Other Python Clients

*   [Generated `lakefs-sdk`](./python/generated_sdk.md)
*   [`lakefs-spec` (for fsspec compatible libraries)](./python/lakefs_spec.md)
*   [`boto3` (S3 Gateway interaction)](./python/boto3_usage.md)
