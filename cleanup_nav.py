# Read the mkdocs.yml content
with open("docs/mkdocs.yml", "r") as f:
    content = f.read()

# Remove trailing backslashes from lines within the nav structure
# This is a common artifact when echoing multi-line shell variables into a file
import re
content = re.sub(r"\\$", "", content, flags=re.MULTILINE)

# Write the cleaned content back
with open("docs/mkdocs.yml", "w") as f:
    f.write(content)

print("Cleaned up trailing backslashes in docs/mkdocs.yml.")
