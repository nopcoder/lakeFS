---
layout: default
title: Using Python with lakeFS
description: Guide to using Python for interacting with lakeFS, covering official SDKs and other integration methods.
parent: Using lakeFS with...
nav_order: 30 # Keep similar nav_order or adjust as needed
has_children: false # Assuming this page itself won't have children pages in this structure
---

# Using Python with lakeFS

Python is a first-class citizen for interacting with lakeFS, offering several ways to integrate lakeFS into your Python applications and data workflows. Whether you prefer a high-level idiomatic Python experience or direct access to the lakeFS API, there's a client that fits your needs.

## Python Integration Options

lakeFS offers a few ways to work with Python:

1.  **High-Level `lakefs` SDK (Recommended):** Provides a user-friendly, idiomatic Python interface for common lakeFS operations like managing repositories, branches, objects, and performing version control actions. This is generally the recommended client for most use cases.
2.  **Generated `lakefs-sdk`:** A Python client library that is auto-generated directly from the lakeFS OpenAPI specification. It provides comprehensive coverage of the lakeFS API, suitable for developers who need access to specific low-level API endpoints or prefer a direct mapping to API calls.
3.  **`lakefs-spec`:** An fsspec-compatible file system implementation for lakeFS. This is ideal for integrating with popular Python data libraries like Pandas, Dask, and others that leverage `fsspec` for data access.
4.  **`boto3` (via S3 Gateway):** lakeFS exposes an S3-compatible gateway. You can use `boto3`, the AWS SDK for Python, to interact with objects in lakeFS as if they were in an S3 bucket. This is useful for leveraging existing S3-based tools and workflows.

**Note on Legacy Clients:**
Previous versions of this documentation and some older community examples might refer to using libraries like `bravado` to dynamically generate a client from a `swagger.json` file. While this approach works, the officially supported and recommended clients are the High-Level `lakefs` SDK and the generated `lakefs-sdk` mentioned above, as they offer better support, performance, and an improved developer experience.

## Using the High-Level `lakefs` SDK (Recommended)

The High-Level `lakefs` SDK is designed to provide a Pythonic and intuitive way to interact with lakeFS. It simplifies common versioning operations and data management tasks, making it easier to integrate lakeFS into your Python scripts and applications.

**Why use the `lakefs` SDK?**

*   **Ease of Use:** Offers a higher-level abstraction over the lakeFS API, with methods that are more aligned with typical Python development patterns.
*   **Idiomatic Python:** Designed to feel natural for Python developers.
*   **Focus on Common Operations:** Streamlines the most frequent tasks like creating branches, committing data, uploading/downloading objects, and merging.
*   **Reduced Boilerplate:** Simplifies authentication and client setup.

### Installation

Install the Python client using pip:

```shell
pip install lakefs
```

### Initialization

The High-Level SDK can often discover lakeFS server settings and credentials if you have `lakectl` configured locally.

**Default Client (with `lakectl` configured):**

If `lakectl` is configured, you can often use the SDK functions directly without explicit client instantiation. Operations will use the settings from your `lakectl` configuration.

```python
import lakefs

# Example: List repositories (if lakectl is configured)
try:
    for repo in lakefs.repositories():
        print(f"Repository: {repo.id}, Default Branch: {repo.default_branch}, Storage: {repo.storage_namespace}")
except Exception as e:
    print(f"Could not list repositories (ensure lakectl is configured or instantiate client explicitly): {e}")

```

**Explicit Client Instantiation:**

For more control, or when `lakectl` is not configured, you can explicitly instantiate a `Client`:

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
    # You can now pass this client object to various lakeFS operations if needed,
    # or use lakefs module functions which will attempt to use this client if it's the last one created.
    # For clarity in examples, we'll often show direct use of lakefs.repository, lakefs.branch etc.
    # which can implicitly use a configured client or discover settings.
    print("lakeFS client initialized successfully (explicitly).")

    # Test with the explicit client by passing it if necessary, though often not if default is set by instantiation.
    # For example, to be absolutely sure a specific client is used:
    # specific_repo = lakefs.Repository("my-repo-id", client=client)

except Exception as e:
    print(f"Error initializing lakeFS client: {e}")
```

**Using TLS with a Custom CA:**

If your lakeFS server uses TLS with a Certificate Authority (CA) not trusted by your host's default trust store, you can specify a CA certificate bundle:

```python
from lakefs.client import Client

# client = Client(
#     host="https://lakefs.example.com", # Note: HTTPS
#     username="YOUR_ACCESS_KEY_ID",
#     password="YOUR_SECRET_ACCESS_KEY",
#     ssl_ca_cert="/path/to/your/ca_bundle.pem"
# )
```
*(Note: This example is commented out as it requires a real path)*

**Disabling SSL Verification (for testing ONLY, NOT recommended for production):**

For local testing against a lakeFS instance with a self-signed certificate, you might need to disable SSL verification. **This is insecure and should NOT be used in production environments.**

```python
from lakefs.client import Client

# client = Client(
#     host="https://localhost:8000", # Or your local HTTPS endpoint
#     username="YOUR_ACCESS_KEY_ID",
#     password="YOUR_SECRET_ACCESS_KEY",
#     verify_ssl=False # DANGER: Insecure, for testing only
# )
```
*(Note: This example is commented out as it's a dangerous setting)*


**Using a Proxy:**

If you need to connect to lakeFS through a proxy server:

```python
from lakefs.client import Client

# client = Client(
#     host="http://lakefs.example.com:8000",
#     username="YOUR_ACCESS_KEY_ID",
#     password="YOUR_SECRET_ACCESS_KEY",
#     proxy="http://your-proxy-server:proxy-port"
# )
```
*(Note: This example is commented out as it requires a real proxy URL)*

---
*Next, we will cover Core Operations.*

### Core Operations

Let's explore some common operations using the `lakefs` SDK. These examples assume you have configured your client (either implicitly via `lakectl` or by explicit instantiation as shown previously).

#### Repositories

Repositories are the top-level containers for your data in lakeFS.

**1. Creating a Repository:**

To create a new repository, you need to specify its name and the underlying storage namespace (e.g., an S3 bucket path).

```python
import lakefs
import datetime

# Assumes client is configured (e.g., via lakectl or explicit Client instantiation)
repo_name = f"my-new-repo-{int(datetime.datetime.now().timestamp())}" # Unique repo name
storage_namespace = f"s3://your-bucket-name/lakefs-data/{repo_name}" # Replace with your actual S3 bucket and desired path

try:
    repo = lakefs.repository(repo_name).create(
        storage_namespace=storage_namespace,
        default_branch="main",  # Optional: specify default branch name
        include_samples=False    # Optional: whether to include sample data
    )
    print(f"Repository '{repo.id}' created successfully.")
    print(f"  Default branch: {repo.default_branch}")
    print(f"  Storage namespace: {repo.storage_namespace}")
    print(f"  Creation date: {datetime.datetime.fromtimestamp(repo.creation_date)}")
except lakefs.exceptions.LakeFSApiException as e:
    if e.status_code == 409: # Conflict, repository already exists
        print(f"Repository '{repo_name}' already exists.")
        repo = lakefs.repository(repo_name) # Get a reference to the existing repo
    else:
        print(f"Error creating repository '{repo_name}': {e}")
except Exception as e:
    print(f"An unexpected error occurred: {e}")
```

**2. Getting a Repository Reference:**

If a repository already exists, you can get a reference to it.

```python
import lakefs

repo_name = "my-new-repo" # Or the name of your existing repo from the previous step
# Ensure this repo_name matches one you expect to exist for this example to work.
# If running this script multiple times, use the timestamped name from the create step,
# or a known existing repository.

try:
    existing_repo = lakefs.repository(repo_name)
    # You might need to call a method to confirm existence or fetch details if needed,
    # as just creating the object doesn't hit the server until an action is performed.
    # For example, to get its properties:
    properties = existing_repo.properties()
    print(f"Successfully got reference to repository '{properties.id}'.")
    print(f"  Default branch: {properties.default_branch}")
    print(f"  Storage namespace: {properties.storage_namespace}")
except lakefs.exceptions.NotFoundException:
    print(f"Repository '{repo_name}' not found.")
except Exception as e:
    print(f"Error getting repository '{repo_name}': {e}")
```

**3. Listing Repositories:**

You can list all repositories accessible with your credentials.

```python
import lakefs
import datetime # Ensure datetime is imported if you use it here for creation_date formatting

try:
    print("\nListing all repositories:")
    count = 0
    for repo_props in lakefs.repositories(): # repo_props is of type RepositoryProperties
        print(f"- ID: {repo_props.id}")
        print(f"  Default Branch: {repo_props.default_branch}")
        print(f"  Storage Namespace: {repo_props.storage_namespace}")
        print(f"  Creation Date: {datetime.datetime.fromtimestamp(repo_props.creation_date)}")
        count += 1
    if count == 0:
        print("No repositories found.")
except Exception as e:
    print(f"Error listing repositories: {e}")
```

**4. Deleting a Repository:** (Use with caution!)

```python
import lakefs

repo_to_delete_name = "my-new-repo" # Replace with the name of the repo you want to delete
# Ensure this is a repository you intend to delete, perhaps one created by this script.

# try:
#     repo_to_delete = lakefs.repository(repo_to_delete_name)
#     repo_to_delete.delete()
#     print(f"Repository '{repo_to_delete_name}' deleted successfully.")
# except lakefs.exceptions.NotFoundException:
#     print(f"Repository '{repo_to_delete_name}' not found for deletion.")
# except Exception as e:
#     print(f"Error deleting repository '{repo_to_delete_name}': {e}")
```
*(Note: Delete operation is commented out for safety in documentation examples)*


#### Branches

Branches allow you to create isolated environments for your work, enabling experimentation and development without affecting the main line of data.

**1. Creating a Branch:**

Branches are created from a source reference, which can be another branch name or a commit ID.

```python
import lakefs

repo_name = "my-new-repo" # Use a known existing repository
# If using the timestamped name from the create step, ensure it's set here.
new_branch_name = "feature-experiment-1"
source_branch_name = "main"

try:
    repo = lakefs.repository(repo_name)
    new_branch = repo.branch(new_branch_name).create(source_reference=source_branch_name)
    print(f"Branch '{new_branch.id}' created successfully in repository '{repo.id}' from source '{source_branch_name}'.")
    # The 'id' of the branch object is its name.
    # The 'new_branch' object is a Ref, you can get its commit:
    # commit = new_branch.get_commit()
    # print(f"  HEAD commit of new branch: {commit.id}")
except lakefs.exceptions.NotFoundException:
    print(f"Repository '{repo_name}' or source branch '{source_branch_name}' not found.")
except lakefs.exceptions.LakeFSApiException as e:
    if e.status_code == 409: # Conflict, branch already exists
        print(f"Branch '{new_branch_name}' already exists in repository '{repo_name}'.")
    else:
        print(f"Error creating branch: {e}")
except Exception as e:
    print(f"An unexpected error occurred while creating branch: {e}")
```

**2. Getting a Branch Reference:**

```python
import lakefs

repo_name = "my-new-repo"
branch_name = "feature-experiment-1" # Or "main"

try:
    branch_ref = lakefs.repository(repo_name).branch(branch_name)
    # To verify it exists and get its head commit:
    commit = branch_ref.get_commit()
    print(f"Successfully got reference to branch '{branch_ref.id}'.")
    print(f"  HEAD commit ID: {commit.id}")
    print(f"  Committer: {commit.committer}")
    print(f"  Message: {commit.message}")
except lakefs.exceptions.NotFoundException:
    print(f"Branch '{branch_name}' or repository '{repo_name}' not found.")
except Exception as e:
    print(f"Error getting branch '{branch_name}': {e}")

```

**3. Listing Branches:**

```python
import lakefs

repo_name = "my-new-repo"

try:
    repo = lakefs.repository(repo_name)
    print(f"\nListing branches for repository '{repo.id}':")
    count = 0
    for branch in repo.branches(): # branch is a Ref object
        print(f"- Branch ID (Name): {branch.id}")
        # head_commit = branch.get_commit() # This would be an extra API call per branch
        # print(f"  HEAD Commit ID: {head_commit.id}")
        count += 1
    if count == 0:
        print("No branches found in this repository.")
except lakefs.exceptions.NotFoundException:
    print(f"Repository '{repo_name}' not found for listing branches.")
except Exception as e:
    print(f"Error listing branches: {e}")
```

**4. Deleting a Branch:** (Use with caution!)

```python
import lakefs

repo_name = "my-new-repo"
branch_to_delete_name = "feature-experiment-1"

# try:
#     repo = lakefs.repository(repo_name)
#     branch_to_delete = repo.branch(branch_to_delete_name)
#     branch_to_delete.delete()
#     print(f"Branch '{branch_to_delete_name}' deleted successfully from repository '{repo.id}'.")
# except lakefs.exceptions.NotFoundException:
#     print(f"Branch '{branch_to_delete_name}' or repository '{repo_name}' not found for deletion.")
# except Exception as e:
#     print(f"Error deleting branch '{branch_to_delete_name}': {e}")

```
*(Note: Delete operation is commented out for safety in documentation examples)*

---
*Next, we will cover Objects (I/O).*
