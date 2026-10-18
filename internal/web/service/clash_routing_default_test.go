package service

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agassiz/3x-ui/v3/internal/util/common"
)

// A default that is not a remote URL is merged as inline YAML, and one that names
// a moved template 404s; either silently drops the routing of every fresh install.
func TestDefaultClashRoutingSourceIsTheShippedTemplate(t *testing.T) {
	raw := defaultValueMap["subClashRules"]
	source, remote, err := common.ParseRemoteRoutingURL(raw)
	if err != nil || !remote {
		t.Fatalf("default subClashRules %q is not a remote routing URL (remote=%v, err=%v)", raw, remote, err)
	}
	u, err := url.Parse(source)
	if err != nil {
		t.Fatalf("parse %q: %v", source, err)
	}
	const branchPrefix = "/agassiz/3x-ui/main/"
	repoPath, ok := strings.CutPrefix(u.Path, branchPrefix)
	if u.Host != "raw.githubusercontent.com" || !ok {
		t.Fatalf("default %q does not point into this fork's main branch", source)
	}
	if _, err := os.Stat(filepath.Join("..", "..", "..", filepath.FromSlash(repoPath))); err != nil {
		t.Fatalf("default points at %s, which is not in the repository: %v", repoPath, err)
	}
}
