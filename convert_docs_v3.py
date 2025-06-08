# Python script to perform content conversion (Version 3)
import os
import re

# Define the root directory for documentation files
docs_root = "docs/docs_mkdocs"

# Regex for Jekyll links: [text]({% link path/to/file.md %})
jekyll_link_regex_v1 = re.compile(r"\[([^\]]+)\]\(\s*{%\s*link\s+([^\s%}]+)\s*%}\s*\)")
# Regex for Jekyll links: {% link path/to/file.md %}
jekyll_link_regex_v2 = re.compile(r"{%\s*link\s+([^\s%}]+)\s*%}")

# Regex for Jekyll includes
known_jekyll_includes_filenames = [ # Just the filenames
    "toc.html", "toc_2-3.html", "toc_2-4.html",
    "head.html", "footer.html", "header_menu.html", "nav.html",
    "mermaid_setup.html",
    "gtag_frame.html",
    "swagger.html",
    "authorization.html",
    "vendor/anchor_headings.html",
    "setup.md"
]
jekyll_include_regex = re.compile(r"{%\s*(?:include|include_relative)\s+([^\s%}]+)\s*%}")

# Regex for Jekyll callouts
# Handles {: .note }, {: .warning }, and {: .warning title="Custom Title" }
jekyll_callout_regex = re.compile(
    r"^\s*{\:\s*\.(note|warning|tip|danger|info|success)\s*" # type
    r"(?:title=\s*\"([^\"]*)\")?\s*}", # optional title="xxx"
    re.MULTILINE
)

# Default admonition types
callout_mapping = {
    "note": "!!! note",
    "warning": "!!! warning",
    "tip": "!!! tip",
    "danger": "!!! danger",
    "info": "!!! info",
    "success": "!!! success",
}

# Predefined titles for specific callout types (used if no title is in the tag)
default_callout_titles = {
    "warning": "⚠️ Warning ⚠️"
}

def convert_callouts_replacement_function(match):
    callout_type = match.group(1)
    custom_title = match.group(2) # This is the title captured from title="..."

    base_admonition = callout_mapping.get(callout_type, f"!!! {callout_type}")

    title_to_use = custom_title # Use title from tag if present
    if not title_to_use:
        title_to_use = default_callout_titles.get(callout_type) # Else, use predefined default

    if title_to_use:
        return f'{base_admonition} "{title_to_use}"'
    else:
        return base_admonition

def filter_includes_replacement_function(match):
    include_param = match.group(1).strip()
    # Check if the filename part of include_param is one of the known layout/structural includes
    include_filename = os.path.basename(include_param)
    if include_filename in known_jekyll_includes_filenames:
        return "" # Remove this include

    # If it's not a known structural include, convert to a comment placeholder
    return f"<!-- Jekyll include: {include_param} needs review -->"


# --- Modifications for constrained execution ---
files_to_process_for_verification = [
    "docs/docs_mkdocs/index.md", # General, might have basic includes removed
    "docs/docs_mkdocs/howto/deploy/aws.md", # Had setup.md include
    "docs/docs_mkdocs/project/docs/callouts.md", # For callout testing
    "docs/docs_mkdocs/howto/garbage-collection/index.md" # Likely has includes that are NOT in known_jekyll_includes_filenames
]
created_files_for_verification = []
# --- End of modifications ---

print(f"Starting content conversion (v3 - verification mode)...")

# --- Modified loop for constrained execution ---
for file_path_to_process in files_to_process_for_verification:
    if not os.path.exists(file_path_to_process):
        print(f"Target file for conversion not found: {file_path_to_process}")
        continue

    converted_file_path = file_path_to_process + ".v3.converted"
    original_content = ""
    print(f"Processing for verification: {file_path_to_process} -> {converted_file_path}")
    try:
        with open(file_path_to_process, 'r', encoding='utf-8') as f_read:
            original_content = f_read.read()

        current_content = original_content

        current_content = jekyll_link_regex_v1.sub(r"[\1](\2)", current_content)
        current_content = jekyll_link_regex_v2.sub(r"\1", current_content)
        current_content = jekyll_include_regex.sub(filter_includes_replacement_function, current_content)
        current_content = jekyll_callout_regex.sub(convert_callouts_replacement_function, current_content)

        with open(converted_file_path, 'w', encoding='utf-8') as f_write:
            f_write.write(current_content)
        created_files_for_verification.append(converted_file_path)
        if current_content == original_content:
            print(f"Created (no changes): {converted_file_path}")
        else:
            print(f"Created (modified): {converted_file_path}")

    except Exception as e:
        print(f"Error processing file {file_path_to_process}: {e}")
# --- End of modified loop ---

print(f"Content conversion (v3 - verification mode) finished.")
if created_files_for_verification:
    print(f"Verification files created: {created_files_for_verification}")
else:
    print("No verification files were created.")
