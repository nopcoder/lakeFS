# Python script to perform content conversion (modified for verification)
import os
import re

# Define the root directory for documentation files
docs_root = "docs/docs_mkdocs"

# Regex for Jekyll links: [text]({% link path/to/file.md %})
jekyll_link_regex_v1 = re.compile(r"\[([^\]]+)\]\(\s*{%\s*link\s+([^\s%}]+)\s*%}\s*\)")
jekyll_link_regex_v2 = re.compile(r"{%\s*link\s+([^\s%}]+)\s*%}")

known_jekyll_includes = [
    "toc.html", "toc_2-3.html", "toc_2-4.html",
    "head.html", "footer.html", "header_menu.html", "nav.html",
    "mermaid_setup.html",
    "gtag_frame.html",
    "swagger.html",
    "authorization.html",
    "vendor/anchor_headings.html",
    "setup.md"
]
jekyll_include_regex = re.compile(r"{%\s*include\s+(" + "|".join(re.escape(inc) for inc in known_jekyll_includes) + r")\s*%}")

jekyll_callout_simple_regex = re.compile(r"^{\:\s*\.(note|warning|tip|danger|info|success)\s*}", re.MULTILINE)

callout_mapping = {
    "note": "!!! note",
    "warning": "!!! warning \"⚠️ Warning ⚠️\"",
    "tip": "!!! tip",
    "danger": "!!! danger",
    "info": "!!! info",
    "success": "!!! success",
}

def convert_callouts(content):
    def replace_callout(match):
        callout_type = match.group(1)
        return callout_mapping.get(callout_type, f"!!! {callout_type}")
    return jekyll_callout_simple_regex.sub(replace_callout, content)

# Files to process for verification
files_to_process = [
    "docs/docs_mkdocs/index.md",
    "docs/docs_mkdocs/howto/deploy/aws.md"
]
created_files_for_verification = []

for file_path in files_to_process:
    if os.path.exists(file_path):
        original_content = ""
        converted_file_path = file_path + ".converted"
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                original_content = f.read()

            content = original_content

            content = jekyll_link_regex_v1.sub(r"[\1](\2)", content)
            content = jekyll_link_regex_v2.sub(r"\1", content)
            content = jekyll_include_regex.sub("", content)
            content = convert_callouts(content)

            if content != original_content:
                with open(converted_file_path, 'w', encoding='utf-8') as f:
                    f.write(content)
                print(f"Created verification file: {converted_file_path}")
                created_files_for_verification.append(converted_file_path)
            else:
                # If content is unchanged, still write it to show it was processed
                with open(converted_file_path, 'w', encoding='utf-8') as f:
                    f.write(content)
                print(f"Created verification file (content unchanged): {converted_file_path}")
                created_files_for_verification.append(converted_file_path)


        except Exception as e:
            print(f"Error processing file {file_path} for verification: {e}")
    else:
        print(f"Target file for conversion not found: {file_path}")

if not created_files_for_verification:
    print("No verification files were created.")
else:
    print(f"Finished creating verification files: {created_files_for_verification}")
