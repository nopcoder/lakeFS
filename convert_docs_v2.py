# Python script to perform content conversion (Version 2)
import os
import re

# Define the root directory for documentation files
docs_root = "docs/docs_mkdocs"

# Regex for Jekyll links: [text]({% link path/to/file.md %})
jekyll_link_regex_v1 = re.compile(r"\[([^\]]+)\]\(\s*{%\s*link\s+([^\s%}]+)\s*%}\s*\)")
# Regex for Jekyll links: {% link path/to/file.md %}
jekyll_link_regex_v2 = re.compile(r"{%\s*link\s+([^\s%}]+)\s*%}")

# Regex for Jekyll includes
known_jekyll_includes = [
    "toc.html", "toc_2-3.html", "toc_2-4.html",
    "head.html", "footer.html", "header_menu.html", "nav.html",
    "mermaid_setup.html",
    "gtag_frame.html",
    "swagger.html",
    "authorization.html",
    "vendor/anchor_headings.html",
    "setup.md" # from howto/deploy/includes
]
# This regex will look for {% include filename %} or {% include path/to/filename %}
# It's simplified; complex Liquid logic within include tags won't be parsed.
jekyll_include_regex = re.compile(r"{%\s*(?:include|include_relative)\s+([^\s%}]+)\s*%}")

# Regex for Jekyll callouts (simple version)
# Catches {: .note }, {: .warning }, etc. at the beginning of a line
# Now also captures optional title="Custom Title"
jekyll_callout_simple_regex = re.compile(r"^\s*{\:\s*\.(note|warning|tip|danger|info|success)\s*(?:title=\s*\"([^\"]*)\")?\s*}", re.MULTILINE)


callout_mapping = {
    "note": "!!! note",
    "warning": "!!! warning", # Default title for warning
    "tip": "!!! tip",
    "danger": "!!! danger",
    "info": "!!! info",
    "success": "!!! success",
}

# Titles for specific callouts from _config.yml or common usage
callout_titles = {
    "warning": "⚠️ Warning ⚠️" # From original _config.yml
}

def convert_callouts(content):
    def replace_callout(match):
        callout_type = match.group(1)
        custom_title = match.group(2) # Might be None if title attribute wasn't present

        base_admonition = callout_mapping.get(callout_type, f"!!! {callout_type}")

        # Use custom title from regex if present, else use predefined title from callout_titles, else no title
        title_to_use = custom_title if custom_title else callout_titles.get(callout_type)

        if title_to_use:
            return f'{base_admonition} "{title_to_use}"'
        else:
            return base_admonition # No title, just !!! type

    return jekyll_callout_simple_regex.sub(replace_callout, content)

def filter_includes(match):
    include_param = match.group(1).strip()
    # Check if the include_param is one of the known layout/structural includes
    # by checking if the *end* of the include_param matches a known_include.
    for known_include_file in known_jekyll_includes:
        if include_param.endswith(known_include_file):
            return "" # Remove this include by returning an empty string
    # If it's not a known structural include (based on its filename part), keep it.
    return match.group(0)


# --- Modifications for constrained execution ---
# Process only a subset of files and write to .converted for verification
files_to_process_for_verification = [
    "docs/docs_mkdocs/index.md",
    "docs/docs_mkdocs/howto/deploy/aws.md",
    "docs/docs_mkdocs/quickstart/index.md",
    "docs/docs_mkdocs/project/docs/callouts.md" # A file that specifically uses callouts with titles
]
created_files_for_verification = []
# --- End of modifications ---

print(f"Starting content conversion (v2 - verification mode)...")

# --- Modified loop for constrained execution ---
for file_path in files_to_process_for_verification:
    if not os.path.exists(file_path):
        print(f"Target file for conversion not found: {file_path}")
        continue

    converted_file_path = file_path + ".v2.converted"
    original_content = ""
    print(f"Processing for verification: {file_path} -> {converted_file_path}")
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            original_content = f.read()

        content = original_content

        content = jekyll_link_regex_v1.sub(r"[\1](\2)", content)
        content = jekyll_link_regex_v2.sub(r"\1", content)
        content = jekyll_include_regex.sub(filter_includes, content)
        content = convert_callouts(content)

        # Always write, even if no changes, to confirm processing
        with open(converted_file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        created_files_for_verification.append(converted_file_path)
        if content == original_content:
            print(f"Created (no changes): {converted_file_path}")
        else:
            print(f"Created (modified): {converted_file_path}")

    except Exception as e:
        print(f"Error processing file {file_path}: {e}")
# --- End of modified loop ---

print(f"Content conversion (v2 - verification mode) finished.")
if created_files_for_verification:
    print(f"Verification files created: {created_files_for_verification}")
else:
    print("No verification files were created.")

# The original script's verification print loop is not needed here,
# as we will use read_files() in the next step.
