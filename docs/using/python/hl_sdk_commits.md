---
layout: default
title: High-Level SDK - Committing Changes
parent: Using Python with lakeFS - Overview
grand_parent: Using lakeFS with...
nav_order: 5
---

Committing changes in lakeFS is a fundamental version control operation. It creates an immutable snapshot of the data and metadata in a branch at a specific point in time. Each commit is identified by a unique ID and can include descriptive metadata.

These examples assume you have already [initialized the lakeFS client](./hl_sdk_overview.md#initialization) and are working with an existing repository and branch where changes (like [object uploads](./hl_sdk_objects.md)) have been made.

Let's assume `repo_name` and `branch_name` variables hold the names of your existing repository and branch:
```python
# Placeholder for setup - ensure these variables are set in your actual script
repo_name = "my-sdk-repo" # Replace with your repository name
branch_name = "main"      # Replace with your branch name where changes were made
```

## 1. Committing Changes

After you've made changes to a branch (e.g., uploaded, modified, or deleted objects), you can commit these changes. A commit message is required, and you can optionally include key-value metadata.

```python
import lakefs
import datetime # For formatting commit timestamps

# repo_name and branch_name should be set
# Ensure some changes have been made to this branch, e.g., by running examples from Objects (I/O) section.

commit_message = "Added initial sensor data and processing scripts"
commit_metadata = {
    "data_source": "iot_stream_processor_v1.2",
    "processed_by": "data_pipeline_user_A",
    "quality_checked": "true"
}

try:
    print(f"Attempting to commit changes on branch '{branch_name}' in repository '{repo_name}'...")
    branch_ref = lakefs.repository(repo_name).branch(branch_name)
    
    # Optional: Check for uncommitted changes first.
    # Note: .uncommitted() returns a generator. Convert to list to check its length or iterate multiple times.
    uncommitted_changes = list(branch_ref.uncommitted()) 
    if not uncommitted_changes:
        print(f"No uncommitted changes found on branch '{branch_name}'. Nothing to commit.")
    else:
        print("Found uncommitted changes:")
        for change in uncommitted_changes:
            print(f"  - Path: {change.path}, Type: {change.type}, Size (bytes): {change.size_bytes if change.size_bytes is not None else 'N/A'}")

        # Perform the commit
        # The .commit() method is called on the Branch object.
        commit_result = branch_ref.commit(
            message=commit_message,
            metadata=commit_metadata
        )
        
        print(f"Commit successful on branch '{branch_name}'.")
        print(f"  Commit ID: {commit_result.id}")
        print(f"  Timestamp: {datetime.datetime.fromtimestamp(commit_result.creation_date)}")
        print(f"  Message: {commit_result.message}")
        print(f"  Metadata: {commit_result.metadata}")
        print(f"  Parents: {commit_result.parents}") # List of parent commit IDs

        # Verify no uncommitted changes remain (should be empty if commit was successful)
        # remaining_changes = list(branch_ref.uncommitted())
        # if not remaining_changes:
        #     print("Successfully committed. No uncommitted changes remain on the branch.")
        # else:
        #     print(f"Warning: {len(remaining_changes)} uncommitted changes still exist after commit. This is unexpected.")

except lakefs.exceptions.NotFoundException:
    print(f"Error: Repository '{repo_name}' or branch '{branch_name}' not found.")
except lakefs.exceptions.LakeFSApiException as e:
    # lakeFS API might return 400 if there's nothing to commit (though SDK might handle this)
    if e.status_code == 400 and "nothing to commit" in str(e.body).lower():
        print(f"Nothing to commit on branch '{branch_name}'. The branch is clean.")
    else:
        print(f"API error during commit: {e} (Status: {e.status_code}, Body: {e.body})")
except Exception as e:
    print(f"An unexpected error occurred during commit: {e}")
```

## 2. Listing Commits (Commit Log)

You can retrieve the history of commits for any branch, tag, or even a specific commit ID (to see its parents).

```python
import lakefs
import datetime

# repo_name and branch_name (or any ref like a commit ID or tag) should be set
ref_to_log = branch_name # Could also be a specific commit_result.id from above, or a tag name

try:
    print(f"\nCommit log for reference '{ref_to_log}' in repository '{repo_name}':")
    
    # Get a Ref object for the branch/commit/tag
    ref_object = lakefs.repository(repo_name).ref(ref_to_log) 
    
    commit_count = 0
    # ref_object.log() returns a generator of Commit objects.
    # max_amount limits the number of commits returned (useful for long histories).
    # Set limit_type to 'TIME' to limit by date range (not shown here).
    for commit_item in ref_object.log(max_amount=10): 
        print(f"- Commit ID: {commit_item.id}")
        print(f"  Committer: {commit_item.committer}")
        print(f"  Message: {commit_item.message}")
        print(f"  Timestamp: {datetime.datetime.fromtimestamp(commit_item.creation_date)}")
        print(f"  Metadata: {commit_item.metadata}")
        print(f"  Parents: {commit_item.parents}")
        print(f"  Generation: {commit_item.generation}") # Useful for understanding commit history depth
        # print(f"  MetaRange ID: {commit_item.meta_range_id}") # Internal detail
        commit_count += 1
        
    if commit_count == 0:
        print(f"No commits found for reference '{ref_to_log}'. This might indicate an empty repository or an incorrect reference.")

except lakefs.exceptions.NotFoundException:
    print(f"Error: Repository '{repo_name}' or reference '{ref_to_log}' not found.")
except lakefs.exceptions.LakeFSApiException as e:
    print(f"API error listing commits: {e}")
except Exception as e:
    print(f"An unexpected error occurred while listing commits: {e}")
```

## 3. Diffing (Uncommitted Changes on a Branch)

To see what changes (added, removed, modified files) are currently staged on a branch *before* committing them. This is also known as checking the "dirty" state of a branch.

```python
import lakefs

# repo_name and branch_name should be set

# Before running this, it's best to make some uncommitted changes to the branch.
# For example, upload a new temporary file:
# try:
#    temp_file_path = "temp_for_diff_example.txt"
#    lakefs.repository(repo_name).branch(branch_name).object(temp_file_path).upload(
#        f"This is a temporary file created at {datetime.datetime.now()}"
#    )
#    print(f"Uploaded temporary file '{temp_file_path}' to branch '{branch_name}' for diff example.")
# except Exception as e:
#    print(f"Error uploading temporary file for diff example: {e}")


try:
    print(f"\nChecking for uncommitted changes on branch '{branch_name}' in repository '{repo_name}':")
    branch_ref = lakefs.repository(repo_name).branch(branch_name)
    
    # branch_ref.uncommitted() returns a generator of Diff objects.
    uncommitted_diffs = list(branch_ref.uncommitted()) 
    
    if not uncommitted_diffs:
        print("No uncommitted changes found on this branch. It's clean.")
    else:
        print("Uncommitted changes found:")
        for diff_entry in uncommitted_diffs: 
            print(f"- Path: {diff_entry.path}")
            print(f"  Type of change: {diff_entry.type}") # e.g., 'added', 'removed', 'changed', 'conflict'
            print(f"  Path Type: {diff_entry.path_type}") # e.g., 'object', 'prefix', 'common_prefix'
            if diff_entry.size_bytes is not None:
                print(f"  Size (bytes): {diff_entry.size_bytes}")
            # For 'changed' type, physical_address might show the new underlying storage location
            # if diff_entry.type == 'changed' and diff_entry.physical_address is not None:
            #      print(f"  New Physical Address: {diff_entry.physical_address}")

    # Clean up the temporary file if you created one for this example
    # try:
    #    if temp_file_path: # Check if variable exists
    #        lakefs.repository(repo_name).branch(branch_name).object(temp_file_path).delete()
    #        print(f"Cleaned up temporary file '{temp_file_path}' from branch '{branch_name}'.")
    # except NameError: # temp_file_path might not be defined if setup code was commented out
    #    pass 
    # except Exception as e:
    #    print(f"Error cleaning up temporary file: {e}") # Best effort cleanup

except lakefs.exceptions.NotFoundException:
    print(f"Error: Repository '{repo_name}' or branch '{branch_name}' not found.")
except lakefs.exceptions.LakeFSApiException as e:
    print(f"API error performing diff: {e}")
except Exception as e:
    print(f"An unexpected error occurred during diff: {e}")
```
For diffing committed changes between branches or commits, see the [Merging and Tagging](./hl_sdk_merging_tags.md#2-diffing-between-references-branches-commits-tags) section.
```
