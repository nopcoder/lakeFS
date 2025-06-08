# Python script to perform content conversion (final attempt for full run)
import os
import re
import glob

# Define the root directory for documentation files
docs_root = "docs/docs_mkdocs"

# --- Start of script from previous attempt: convert_docs_v3.py ---
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
jekyll_callout_regex = re.compile(
    r"^\s*{\:\s*\.(note|warning|tip|danger|info|success)\s*" # type
    r"(?:title=\s*\"([^\"]*)\")?\s*}", # optional title="xxx"
    re.MULTILINE
)

callout_mapping = {
    "note": "!!! note",
    "warning": "!!! warning",
    "tip": "!!! tip",
    "danger": "!!! danger",
    "info": "!!! info",
    "success": "!!! success",
}

default_callout_titles = {
    "warning": "⚠️ Warning ⚠️"
}

def convert_callouts_replacement_function(match):
    callout_type = match.group(1)
    custom_title = match.group(2)
    base_admonition = callout_mapping.get(callout_type, f"!!! {callout_type}")
    title_to_use = custom_title if custom_title else default_callout_titles.get(callout_type)
    return f'{base_admonition} "{title_to_use}"' if title_to_use else base_admonition

def filter_includes_replacement_function(match):
    include_param = match.group(1).strip()
    include_filename = os.path.basename(include_param)
    if include_filename in known_jekyll_includes_filenames:
        return ""
    return f"<!-- Jekyll include: {include_param} needs review -->"
# --- End of script from previous attempt ---

modified_files_count = 0
total_files_processed = 0

print(f"Starting content conversion in {docs_root}...")
print("IMPORTANT: This script MUST run over all .md files and modify them IN PLACE.")
print("IMPORTANT: DO NOT run in a 'verification mode' or on a subset of files.")

# Count .md files before processing
initial_md_files = glob.glob(os.path.join(docs_root, "**/*.md"), recursive=True)
print(f"Found {len(initial_md_files)} Markdown files to process.")

for root_dir, _, file_list in os.walk(docs_root):
    for file_name_iter in file_list:
        if file_name_iter.endswith(".md"):
            total_files_processed +=1
            file_path_to_process = os.path.join(root_dir, file_name_iter)

            try:
                with open(file_path_to_process, 'r', encoding='utf-8') as f_read:
                    original_content = f_read.read()

                current_content = original_content
                current_content = jekyll_link_regex_v1.sub(r"[\1](\2)", current_content) # Corrected regex sub
                current_content = jekyll_link_regex_v2.sub(r"\1", current_content)     # Corrected regex sub
                current_content = jekyll_include_regex.sub(filter_includes_replacement_function, current_content)
                current_content = jekyll_callout_regex.sub(convert_callouts_replacement_function, current_content)

                if current_content != original_content:
                    with open(file_path_to_process, 'w', encoding='utf-8') as f_write:
                        f_write.write(current_content)
                    modified_files_count += 1
            except Exception as e:
                # This print might not be visible if the overall session fails due to too many file mods
                print(f"Error processing file {file_path_to_process}: {e}")

# These final print statements are crucial for the subtask, but may not be seen if the run fails early.
print(f"Content conversion finished. Total files processed: {total_files_processed}. Total files modified: {modified_files_count}")

files_needing_review = 0
if total_files_processed > 0 : # Check if any processing happened before attempting to read files again
    for root_dir_v, _, file_list_v in os.walk(docs_root):
        for file_name_v in file_list_v:
            if file_name_v.endswith(".md"):
                file_path_v = os.path.join(root_dir_v, file_name_v)
                try:
                    with open(file_path_v, 'r', encoding='utf-8') as f_v:
                        # Check if file content indicates it was actually modified by *this script run*
                        # This is tricky because changes are rolled back on failure.
                        # If the run fails, this loop will read the *original* files.
                        if "<!-- Jekyll include:" in f_v.read():
                            files_needing_review += 1
                except Exception:
                    pass
    print(f"Verification: Found {files_needing_review} files with 'Jekyll include... needs review' comments (note: may reflect pre-existing state if run failed and rolled back).")

if total_files_processed == len(initial_md_files) and total_files_processed > 0:
    print(f"Successfully processed all {total_files_processed} Markdown files.")
    if modified_files_count > 0:
         print(f"Script logic implies it would have modified {modified_files_count} files in place.")
    else:
        print("Script logic implies no files would have required modification based on its patterns.")
else:
    print(f"Warning: Processed {total_files_processed} files, but initially found {len(initial_md_files)}. Some files may have been missed (or run failed).")
