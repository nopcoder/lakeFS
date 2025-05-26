---
layout: default
title: High-Level SDK - Working with Branches
parent: Using Python with lakeFS - Overview
grand_parent: Using lakeFS with...
nav_order: 3
---

Branches in lakeFS allow you to create isolated versions of your data, enabling experimentation, development, and complex workflows without impacting your main data lines. This section details how to manage branches using the High-Level `lakefs` SDK.

These examples assume you have already [initialized the lakeFS client](./hl_sdk_overview.md#initialization) and are working with an existing repository (refer to the [Repositories guide](./hl_sdk_repositories.md) if needed).

Let's assume `repo_name` variable holds the name of your existing repository:
```python
# Placeholder for setup - ensure this variable is set in your actual script
repo_name = "my-sdk-repo" # Replace with your repository name
```

## 1. Creating a Branch

Branches are typically created from an existing reference, which can be another branch name or a specific commit ID.

```python
import lakefs
import datetime # For HEAD commit timestamp formatting

# repo_name should be set to your existing repository's name
# For example: repo_name = "my-sdk-repo" 

new_branch_name = "feat-data-experiment"
source_branch_name = "main" # Or any existing branch or commit ID

try:
    print(f"Attempting to create branch '{new_branch_name}' from source '{source_branch_name}' in repository '{repo_name}'...")
    repo = lakefs.repository(repo_name)
    
    # The .branch() method returns a Branch object, then .create() is called on it.
    new_branch_ref = repo.branch(new_branch_name).create(source_reference=source_branch_name)
    
    print(f"Branch '{new_branch_ref.id}' created successfully.") # new_branch_ref.id is the branch name
    
    # You can get the HEAD commit of this new branch
    # commit = new_branch_ref.get_commit() # This makes an API call
    # print(f"  HEAD commit ID of '{new_branch_ref.id}': {commit.id}")
    # print(f"  HEAD commit message: {commit.message}")
    # print(f"  HEAD commit timestamp: {datetime.datetime.fromtimestamp(commit.creation_date)}")

except lakefs.exceptions.NotFoundException:
    print(f"Error: Repository '{repo_name}' or source branch/reference '{source_branch_name}' not found.")
except lakefs.exceptions.LakeFSApiException as e:
    if e.status_code == 409: # HTTP 409 Conflict: Branch likely already exists
        print(f"Branch '{new_branch_name}' already exists in repository '{repo_name}'.")
    else:
        print(f"API error creating branch: {e} (Status: {e.status_code}, Body: {e.body})")
except Exception as e:
    print(f"An unexpected error occurred while creating branch: {e}")
```

## 2. Getting a Branch Reference

To work with an existing branch, you obtain a `Branch` (or more generally, a `Ref`) object.

```python
import lakefs
import datetime

# repo_name should be set to your existing repository's name
# branch_to_get_name = "main" # Or "feat-data-experiment" if created above

# For this example, let's use "main"
branch_to_get_name = "main" 

try:
    print(f"Attempting to get a reference to branch '{branch_to_get_name}' in repository '{repo_name}'...")
    repo = lakefs.repository(repo_name)
    branch_ref = repo.branch(branch_to_get_name) # Creates a local Ref object for the branch

    # To confirm it exists on the server and fetch its details (like HEAD commit):
    head_commit = branch_ref.get_commit() # This makes an API call
    
    print(f"Successfully got reference to branch '{branch_ref.id}'.")
    print(f"  HEAD commit ID: {head_commit.id}")
    print(f"  Committer: {head_commit.committer}")
    print(f"  Message: {head_commit.message}")
    print(f"  Timestamp: {datetime.datetime.fromtimestamp(head_commit.creation_date)}")
    print(f"  Metadata: {head_commit.metadata}")

except lakefs.exceptions.NotFoundException:
    print(f"Error: Branch '{branch_to_get_name}' or repository '{repo_name}' not found on the server.")
except lakefs.exceptions.LakeFSApiException as e:
    print(f"API error getting branch '{branch_to_get_name}': {e}")
except Exception as e:
    print(f"An unexpected error occurred while getting branch '{branch_to_get_name}': {e}")
```

## 3. Listing Branches

You can list all branches in a specific repository.

```python
import lakefs

# repo_name should be set to your existing repository's name

try:
    print(f"\nListing branches for repository '{repo_name}':")
    repo = lakefs.repository(repo_name)
    count = 0
    # repo.branches() returns a generator of Ref objects (branches)
    for branch_item in repo.branches(): 
        print(f"- Branch Name (ID): {branch_item.id}")
        # For more details, like the HEAD commit, you'd make an additional call per branch:
        # head_commit = branch_item.get_commit() 
        # print(f"  HEAD Commit ID: {head_commit.id}")
        count += 1
    
    if count == 0:
        print(f"No branches found in repository '{repo_name}'.")

except lakefs.exceptions.NotFoundException:
    print(f"Error: Repository '{repo_name}' not found for listing branches.")
except lakefs.exceptions.LakeFSApiException as e:
    print(f"API error listing branches: {e}")
except Exception as e:
    print(f"An unexpected error occurred while listing branches: {e}")
```

## 4. Deleting a Branch

Deleting a branch removes it from the repository. This operation should be used with care, especially for branches that haven't been merged or backed up.

```python
import lakefs

# repo_name should be set to your existing repository's name
branch_to_delete_name = "feat-data-experiment-to-delete" # Use a specific branch name you intend to delete

# print(f"Attempting to delete branch '{branch_to_delete_name}' from repository '{repo_name}'...")
# print("WARNING: This operation can lead to data loss if the branch is not merged or backed up.")

# try:
#     repo = lakefs.repository(repo_name)
#     branch_to_delete = repo.branch(branch_to_delete_name)
#     branch_to_delete.delete() # Makes an API call to delete the branch
#     print(f"Branch '{branch_to_delete_name}' deleted successfully from repository '{repo.id}'.")

# except lakefs.exceptions.NotFoundException:
#     print(f"Error: Branch '{branch_to_delete_name}' or repository '{repo_name}' not found, cannot delete.")
# except lakefs.exceptions.LakeFSApiException as e:
#     # e.g., if trying to delete the default branch, or insufficient permissions
#     print(f"API error deleting branch '{branch_to_delete_name}': {e} (Status: {e.status_code})")
# except Exception as e:
#     print(f"An unexpected error occurred while deleting branch '{branch_to_delete_name}': {e}")
```
*(Note: The delete operation code is commented out by default in this documentation example to prevent accidental deletion. Ensure the branch name is correct and you intend to delete it before uncommenting.)*
```
