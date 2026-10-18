package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDockerDevelopmentScripts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("development scripts require Bash")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("Bash is unavailable")
	}
	cases := []struct {
		name       string
		script     string
		args       []string
		buildExit  string
		upExit     string
		wantOK     bool
		wantUp     bool
		wantTokens []string
	}{
		{"build failure preserves running container", "docker-dev.sh", nil, "1", "0", false, false, nil},
		{"startup failure is reported", "docker-dev.sh", nil, "0", "1", false, true, nil},
		{"rebuild then start without a second build", "docker-dev.sh", []string{"--rebuild"}, "0", "0", true, true,
			[]string{"<build> <--no-cache> <3xui>", "<up> <-d> <--no-build> <--wait> <--wait-timeout> <60> <3xui>"}},
		{"standalone build uses the same compose file", "docker-build-fast.sh", []string{"--no-cache", "--verbose"}, "0", "0", true, false,
			[]string{"<--progress> <plain>", "<build> <--no-cache> <3xui>"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := filepath.Join(t.TempDir(), "project with spaces")
			if err := os.MkdirAll(project, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"docker-dev.sh", "docker-build-fast.sh", "docker-compose.dev.yml"} {
				source, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(project, name), source, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			mockBin := t.TempDir()
			logPath := filepath.Join(t.TempDir(), "docker.log")
			mockDocker := `#!/bin/sh
printf '%s' "$PWD" >> "$XUI_TEST_DOCKER_LOG"
printf ' <%s>' "$@" >> "$XUI_TEST_DOCKER_LOG"
printf '\n' >> "$XUI_TEST_DOCKER_LOG"
case " $* " in
  *" build "*) exit "$XUI_TEST_BUILD_EXIT" ;;
  *" up "*) exit "$XUI_TEST_UP_EXIT" ;;
esac
exit 0
`
			if err := os.WriteFile(filepath.Join(mockBin, "docker"), []byte(mockDocker), 0o755); err != nil {
				t.Fatal(err)
			}
			args := append([]string{filepath.Join(project, tc.script)}, tc.args...)
			cmd := exec.Command(bash, args...)
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(),
				"PATH="+mockBin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"XUI_TEST_DOCKER_LOG="+logPath,
				"XUI_TEST_BUILD_EXIT="+tc.buildExit,
				"XUI_TEST_UP_EXIT="+tc.upExit,
			)
			output, err := cmd.CombinedOutput()
			if (err == nil) != tc.wantOK {
				t.Fatalf("exit error=%v, want success=%v; output:\n%s", err, tc.wantOK, output)
			}
			logBytes, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			log := string(logBytes)
			if strings.Contains(log, "<down>") || strings.Contains(log, "<stop>") {
				t.Fatalf("script stopped the existing container:\n%s", log)
			}
			buildAt, upAt := strings.Index(log, "<build>"), strings.Index(log, "<up>")
			if buildAt < 0 || (upAt >= 0) != tc.wantUp || (upAt >= 0 && upAt < buildAt) {
				t.Fatalf("unexpected build/start order:\n%s", log)
			}
			for _, token := range append(tc.wantTokens, "<-f> <"+filepath.Join(project, "docker-compose.dev.yml")+">") {
				if !strings.Contains(log, token) {
					t.Fatalf("missing %q in Docker calls:\n%s", token, log)
				}
			}
			for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
				if !strings.HasPrefix(line, project+" ") {
					t.Fatalf("Docker command used the caller's directory: %s", line)
				}
			}
			if !tc.wantOK && strings.Contains(string(output), "开发环境已启动") {
				t.Fatalf("failure was reported as a successful start:\n%s", output)
			}
		})
	}
}
