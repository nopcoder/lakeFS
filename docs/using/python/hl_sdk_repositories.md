---
layout: default
title: High-Level SDK - Working with Repositories
parent: Using Python with lakeFS - Overview
grand_parent: Using lakeFS with...
nav_order: 2
---

Repositories are the top-level containers for your data in lakeFS. This section covers common repository operations using the High-Level `lakefs` SDK.

These examples assume you have already [initialized the lakeFS client](./hl_sdk_overview.md#initialization).

## 1. Creating a Repository

To create a new repository, you need to specify its name and the underlying storage namespace (e.g., an S3 bucket path where lakeFS will store data for this repository).

```python
import lakefs
import datetime # For generating a unique repository name in the example

# Assumes client is configured (e.g., via lakectl or explicit Client instantiation as shown in overview)
repo_name = f"my-sdk-repo-{int(datetime.datetime.now().timestamp())}" # Example of a unique repo name
# Replace 'your-bucket-name' with your actual bucket and desired path structure
storage_namespace = f"s3://your-bucket-name/lakefs-storage/{repo_name}" 

try:
    print(f"Attempting to create repository '{repo_name}' with storage '{storage_namespace}'...")
    # The lakefs.repository() function returns a Repository object.
    # The .create() method is then called on this object.
    repo = lakefs.repository(repo_name).create(
        storage_namespace=storage_namespace,
        default_branch="main",  # Optional: specify the default branch name, defaults to "main"
        include_samples=False   # Optional: set to True to populate with sample data
    )
    print(f"Repository '{repo.id}' created successfully.")
    print(f"  Default branch: {repo.default_branch}")
    print(f"  Storage namespace: {repo.storage_namespace}")
    # repo.creation_date is a Unix timestamp
    print(f"  Creation date: {datetime.datetime.fromtimestamp(repo.creation_date)}")

except lakefs.exceptions.LakeFSApiException as e:
    if e.status_code == 409: # HTTP 409 Conflict indicates it likely already exists
        print(f"Repository '{repo_name}' already exists. To work with it, you can get a reference to it.")
        # Optionally, get a reference if it already exists:
        # repo = lakefs.repository(repo_name) 
    else:
        # Handle other API-related errors
        print(f"Error creating repository '{repo_name}': {e}")
        print(f"  Status Code: {e.status_code}")
        print(f"  Reason: {e.reason}")
        print(f"  Body: {e.body}")
except Exception as e:
    # Handle other unexpected errors
    print(f"An unexpected error occurred: {e}")
```

## 2. Getting a Repository Reference

If a repository already exists, you can get a `Repository` object to interact with it. This does not create the repository if it doesn't exist but gives you a client-side object.

```python
import lakefs

# Use the name of a repository you expect to exist.
# If you just ran the create example, you can use that 'repo_name'.
# For this example, let's assume 'my-sdk-repo' exists from a previous operation.
repo_name_to_get = "my-sdk-repo" # Replace with a known repository name

try:
    print(f"Attempting to get a reference to repository '{repo_name_to_get}'...")
    repo_ref = lakefs.repository(repo_name_to_get)
    
    # The line above creates a local Repository object. 
    # To confirm it exists on the server and fetch its properties, you can call a method:
    properties = repo_ref.properties() # This makes an API call
    
    print(f"Successfully got reference to repository '{properties.id}'.")
    print(f"  Default branch: {properties.default_branch}")
    print(f"  Storage namespace: {properties.storage_namespace}")
    print(f"  Creation date: {datetime.datetime.fromtimestamp(properties.creation_date)}")

except lakefs.exceptions.NotFoundException:
    # This exception is raised if the repository doesn't exist when .properties() or other server interaction is called.
    print(f"Repository '{repo_name_to_get}' not found on the server.")
except lakefs.exceptions.LakeFSApiException as e:
    print(f"API error getting repository '{repo_name_to_get}': {e}")
except Exception as e:
    print(f"An unexpected error occurred while getting repository '{repo_name_to_get}': {e}")
```

## 3. Listing Repositories

You can list all repositories that are accessible with your current lakeFS credentials.

```python
import lakefs
import datetime

try:
    print("\nListing all accessible repositories:")
    count = 0
    # lakefs.repositories() returns a generator of RepositoryProperties objects
    for repo_props in lakefs.repositories(): 
        print(f"- ID: {repo_props.id}")
        print(f"  Default Branch: {repo_props.default_branch}")
        print(f"  Storage Namespace: {repo_props.storage_namespace}")
        print(f"  Creation Date: {datetime.datetime.fromtimestamp(repo_props.creation_date)}")
        count += 1
    
    if count == 0:
        print("No repositories found or accessible with current credentials.")
except lakefs.exceptions.LakeFSApiException as e:
    print(f"API error listing repositories: {e}")
except Exception as e:
    print(f"An unexpected error occurred while listing repositories: {e}")
```

## 4. Deleting a Repository

Deleting a repository is an irreversible operation and should be used with extreme caution.

```python
import lakefs

# IMPORTANT: Replace with the name of a repository you specifically want to delete.
# For safety, this example often uses a uniquely generated name or is commented out.
repo_to_delete_name = "my-sdk-repo-to-delete" 

# print(f"Attempting to delete repository '{repo_to_delete_name}'...")
# print("WARNING: This operation is irreversible and will delete all data in the repository.")

# try:
#     # First, get a reference to the repository
#     repo_to_delete = lakefs.repository(repo_to_delete_name)
#     # Then, call the delete method
#     repo_to_delete.delete()
#     print(f"Repository '{repo_to_delete_name}' deleted successfully.")

# except lakefs.exceptions.NotFoundException:
#     print(f"Repository '{repo_to_delete_name}' not found, cannot delete.")
# except lakefs.exceptions.LakeFSApiException as e:
#     # Handle other API errors, e.g., permission issues
#     print(f"API error deleting repository '{repo_to_delete_name}': {e}")
# except Exception as e:
#     print(f"An unexpected error occurred while deleting repository '{repo_to_delete_name}': {e}")
```
*(Note: The delete operation code is commented out by default in this documentation example to prevent accidental deletion. Uncomment and use with care.)*
```
