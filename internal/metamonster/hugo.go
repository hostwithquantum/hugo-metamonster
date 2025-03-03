package metamonster

import (
	"bufio"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
)

type HugoContent map[string]string

// ListContent lists content from hugo (aka `hugo list all`)
func ListContent(ctx context.Context, site, hugo string) (HugoContent, error) {
	cmd := exec.CommandContext(ctx, hugo, "list", "all")
	cmd.Dir = site

	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer output.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var content = make(HugoContent)

	stdin := bufio.NewScanner(output)
	for stdin.Scan() {
		parts := strings.Split(stdin.Text(), ",")
		// csv header/footer
		if parts[7] == "permalink" {
			continue
		}

		content[parts[7]] = filepath.Join(site, parts[0])
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return content, nil
}
