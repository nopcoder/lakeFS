---
layout: default
title: High-Level SDK - Overview
parent: Using Python with lakeFS - Overview
grand_parent: Using lakeFS with...
nav_order: 1
---

The High-Level `lakefs` SDK is designed to provide a Pythonic and intuitive way to interact with lakeFS. It simplifies common versioning operations and data management tasks, making it easier to integrate lakeFS into your Python scripts and applications.

## Why use the `lakefs` SDK?

*   **Ease of Use:** Offers a higher-level abstraction over the lakeFS API, with methods that are more aligned with typical Python development patterns.
*   **Idiomatic Python:** Designed to feel natural for Python developers.
*   **Focus on Common Operations:** Streamlines the most frequent tasks like creating branches, committing data, uploading/downloading objects, and merging.
*   **Reduced Boilerplate:** Simplifies authentication and client setup.

## Installation

Install the Python client using pip:

```shell
pip install lakefs
```

## Initialization

The High-Level SDK can often discover lakeFS server settings and credentials if you have `lakectl` configured locally.

### Default Client (with `lakectl` configured)

If `lakectl` is configured, you can often use the SDK functions directly without explicit client instantiation. Operations will use the settings from your `lakectl` configuration.

```python
import lakefs

# Example: List repositories (if lakectl is configured)
# This assumes your lakectl is set up to connect to your lakeFS server.
try:
    print("Attempting to list repositories using default client configuration...")
    for repo_item in lakefs.repositories(): # Use a different variable name like repo_item
        print(f"Repository: {repo_item.id}, Default Branch: {repo_item.default_branch}, Storage: {repo_item.storage_namespace}")
    print("Successfully listed repositories or no repositories found.")
except Exception as e:
    print(f"Could not list repositories using default client (ensure lakectl is configured or instantiate client explicitly): {e}")
```

### Explicit Client Instantiation

For more control, or when `lakectl` is not configured (e.g., in a CI/CD environment or a remote notebook), you can explicitly instantiate a `Client`:

```python
from lakefs.client import Client

# Replace with your lakeFS endpoint and credentials
LAKEFS_ENDPOINT = "http://lakefs.example.com:8000" # Or your actual lakeFS server endpoint
LAKEFS_ACCESS_KEY_ID = "YOUR_ACCESS_KEY_ID"
LAKEFS_SECRET_ACCESS_KEY = "YOUR_SECRET_ACCESS_KEY"

try:
    client = Client(
        host=LAKEFS_ENDPOINT,
        username=LAKEFS_ACCESS_KEY_ID,
        password=LAKEFS_SECRET_ACCESS_KEY
    )
    print("lakeFS client initialized successfully (explicitly).")
    # You can now pass this 'client' object to various lakeFS operations if needed,
    # or use lakefs module functions which will attempt to use this client if it's the last one created and active.
    # For clarity in subsequent examples, we'll often show direct use of lakefs.repository(), lakefs.branch(), etc.,
    # which can implicitly use a configured client or discover settings from environment variables or lakectl.
    
    # Example: Test the explicitly configured client
    # print("Listing repositories with the explicit client:")
    # for repo_item in lakefs.repositories(client=client): # Pass client explicitly if needed
    #     print(f"Repository: {repo_item.id}")

except Exception as e:
    print(f"Error initializing lakeFS client: {e}")
```
**Note:** After explicitly creating a `Client` instance, subsequent calls to top-level `lakefs` functions (like `lakefs.repositories()`) in the same Python session might implicitly use the last created client if no other client is actively configured (e.g. via `lakectl` or environment variables recognized by the SDK). For robust behavior, especially if managing multiple client instances or in complex applications, explicitly passing the `client` object to SDK functions that accept it is recommended.


### Using TLS with a Custom CA

If your lakeFS server uses TLS with a Certificate Authority (CA) not trusted by your host's default trust store, you can specify a CA certificate bundle file:

```python
from lakefs.client import Client

# client = Client(
#     host="https://lakefs.example.com", # Note: HTTPS for TLS
#     username="YOUR_ACCESS_KEY_ID",
#     password="YOUR_SECRET_ACCESS_KEY",
#     ssl_ca_cert="/path/to/your/ca_bundle.pem" # Provide the actual path to your PEM file
# )
# print("Client configured for TLS with custom CA.")
```
*(Note: This example is commented out as it requires a real file path and endpoint configuration.)*

### Disabling SSL Verification (for testing ONLY)

For local testing against a lakeFS instance with a self-signed certificate, you might need to disable SSL verification. **This is insecure and should NOT be used in production environments due to security risks like man-in-the-middle attacks.**

```python
from lakefs.client import Client

# client = Client(
#     host="https://localhost:8000", # Or your local HTTPS endpoint with a self-signed cert
#     username="YOUR_ACCESS_KEY_ID",
#     password="YOUR_SECRET_ACCESS_KEY",
#     verify_ssl=False # DANGER: Insecure, for testing local setups only
# )
# print("Client configured with SSL verification disabled (INSECURE).")
```
*(Note: This example is commented out due to its insecure nature.)*


### Using a Proxy

If you need to connect to lakeFS through an HTTP/HTTPS proxy server:

```python
from lakefs.client import Client

# client = Client(
#     host="http://lakefs.example.com:8000",
#     username="YOUR_ACCESS_KEY_ID",
#     password="YOUR_SECRET_ACCESS_KEY",
#     proxy="http://your-proxy-server-address:proxy-port" # Replace with your actual proxy URL
# )
# print("Client configured to use a proxy server.")
```
*(Note: This example is commented out as it requires a real proxy URL.)*
