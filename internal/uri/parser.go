package uri

import (
	"fmt"
	"strings"

	"github.com/treeverse/lakefs/pkg/uri" // Using existing lakeFS URI pkg for base parsing
	"github.com/example/lctl/internal/config" // Corrected lctl's config for CLIContext
)

// LakeFSUri is a wrapper around lakeFS's uri.URI providing additional context
// about how it was parsed or resolved.
type LakeFSUri struct {
	*uri.URI                                  // Embed the original lakeFS URI type
	OriginalInput string                      // The raw string given by the user
	Resolved      bool                        // True if context was used to resolve parts
	// Repository, Ref, Path fields are available directly from the embedded uri.URI
}

// DefaultPathResolutionType helps decide how to treat an ambiguous single-segment path
// when full context (repo and ref) is available.
type DefaultPathResolutionType int

const (
	// DefaultPathIsRef treats an ambiguous path like "mybranch" as a reference.
	DefaultPathIsRef DefaultPathResolutionType = iota
	// DefaultPathIsDir treats an ambiguous path like "mydir" as a directory under the current context.
	DefaultPathIsDir
	// DefaultPathIsObject treats an ambiguous path like "myfile.txt" as an object under the current context.
	DefaultPathIsObject
)

// ResolveLakeFSUri parses a raw path string. If the path is not a full lakeFS URI
// (i.e., doesn't start with "lakefs://"), it attempts to resolve it using the provided CLIContext.
// defaultResType is used to guide resolution when a path could be a ref or an object/directory name.
func ResolveLakeFSUri(rawPath string, cliCtx *config.CLIContext, defaultResType DefaultPathResolutionType) (*LakeFSUri, error) {
	if rawPath == "" && (cliCtx == nil || cliCtx.CurrentRepoURI == "" || cliCtx.CurrentRef == "") {
		return nil, fmt.Errorf("cannot resolve empty path without full repository and ref context")
	}
    if rawPath == "" && cliCtx != nil && cliCtx.CurrentRepoURI != "" && cliCtx.CurrentRef != "" {
        // Empty path with full context means root of the current ref
        repoURI, err := uri.Parse(cliCtx.CurrentRepoURI)
        if err != nil {
            return nil, fmt.Errorf("invalid repository context URI '%s': %w", cliCtx.CurrentRepoURI, err)
        }
        finalURIStr := fmt.Sprintf("lakefs://%s/%s/", repoURI.Repository, cliCtx.CurrentRef)
        parsed, err := uri.Parse(finalURIStr)
        if err != nil {
            return nil, fmt.Errorf("failed to construct URI for root of context: %w", err)
        }
        return &LakeFSUri{URI: parsed, OriginalInput: rawPath, Resolved: true}, nil
    }


	if strings.HasPrefix(rawPath, uri.LakeFSScheme+"://") {
		parsed, err := uri.Parse(rawPath)
		if err != nil {
			return nil, fmt.Errorf("invalid full lakeFS URI '%s': %w", rawPath, err)
		}
		// Ensure path is not empty if it's just repo/branch. Add trailing slash for consistency if it's a "directory like" ref.
        // The lakefs/uri package might already handle this. For example, uri.Path should be "/" if not specified.
        // If parsed.Path == "" and parsed.Ref != "", it could mean the ref itself.
        // If path is truly empty (e.g. lakefs://repo/branch), GetPath() should return "/"
		return &LakeFSUri{URI: parsed, OriginalInput: rawPath, Resolved: false}, nil
	}

	// Not a full URI, try to resolve with context
	if cliCtx == nil {
		return nil, fmt.Errorf("cannot resolve partial path '%s' without CLI context", rawPath)
	}
	if cliCtx.CurrentRepoURI == "" {
		return nil, fmt.Errorf("repository context not set; cannot resolve partial path '%s'", rawPath)
	}

	repoCtxURI, err := uri.Parse(cliCtx.CurrentRepoURI)
	if err != nil {
		return nil, fmt.Errorf("invalid repository context URI '%s': %w", cliCtx.CurrentRepoURI, err)
	}
	resolvedRepo := repoCtxURI.Repository
	resolvedRef := cliCtx.CurrentRef
	resolvedPathPart := rawPath
	wasResolved := true

	// Resolution logic:
	// 1. "ref:/path/to/obj"
	// 2. "/path/to/obj" (needs ref from context)
	// 3. "path/to/obj" or "object" (needs ref from context, path relative to ref root)
	// 4. "refonly" (if defaultResType is DefaultPathIsRef)

	parts := strings.SplitN(rawPath, ":", 2)
	if len(parts) == 2 && !strings.Contains(parts[0], "/") {
		// Case 1: "ref:/path/to/obj" - overrides context ref
		resolvedRef = parts[0]
		resolvedPathPart = parts[1]
	} else {
		// Cases 2, 3, 4 - ref comes from context or is the path itself
		if resolvedRef == "" && defaultResType != DefaultPathIsRef {
             return nil, fmt.Errorf("ref context not set; cannot resolve path '%s' without explicit ref like 'mybranch:/path'", rawPath)
        }
        // If rawPath contains no slashes, and we expect a ref, it could be a ref.
        if !strings.Contains(rawPath, "/") && defaultResType == DefaultPathIsRef {
            resolvedRef = rawPath
            resolvedPathPart = "/" // Default to root of this ref
        } else {
            // Otherwise, rawPath is a path relative to resolvedRef (which must be set from context)
            if resolvedRef == "" { // Should have been caught above if defaultResType wasn't DefaultPathIsRef
                 return nil, fmt.Errorf("ref context not set or path '%s' is ambiguous; specify as 'mybranch:/path' or set ref context", rawPath)
            }
            resolvedPathPart = rawPath // Already set
        }
	}

	// Ensure resolvedPathPart starts with "/" if it's not empty or already starting with "/"
	if resolvedPathPart != "" && !strings.HasPrefix(resolvedPathPart, "/") {
		resolvedPathPart = "/" + resolvedPathPart
	}
    if resolvedPathPart == "" { // If after all logic path is empty, make it "/" for root.
        resolvedPathPart = "/"
    }


	finalURIStr := fmt.Sprintf("%s://%s/%s%s", uri.LakeFSScheme, resolvedRepo, resolvedRef, resolvedPathPart)
	// Normalize trailing slash for "directory like" paths vs "object like" paths
    // If the original input ended with a slash, or if it's a directory by context, preserve/add it.
    // The pkg/uri.Parse should handle this correctly. For instance, path "/" is root.
    // If resolvedPathPart ends up being just "/", it's fine.
    // If resolvedPathPart is like "/foo/" it's also fine.
    // If resolvedPathPart is like "/foo" for an object, also fine.

	parsed, err := uri.Parse(finalURIStr)
	if err != nil {
		return nil, fmt.Errorf("failed to construct resolved URI from ('%s', '%s', '%s'): %w", resolvedRepo, resolvedRef, resolvedPathPart, err)
	}

	return &LakeFSUri{URI: parsed, OriginalInput: rawPath, Resolved: wasResolved}, nil
}
