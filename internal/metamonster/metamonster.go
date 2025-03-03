package metamonster

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

// pageParamKey to stack (optimized) title, description and keywords
const pageParamKey = "meta"

func Update(ctx context.Context, path string, update Metamonster) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	slog.DebugContext(ctx, path)

	page, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read contents: %s", err)
	}

	// Parse the front matter and the content
	var metadata map[string]interface{}
	content, err := frontmatter.Parse(bytes.NewReader(page), &metadata)
	if err != nil {
		return err
	}

	if _, ok := metadata[pageParamKey]; !ok {
		metadata[pageParamKey] = make(map[interface{}]interface{})
	}

	// Add optimized keywords and description
	metadata[pageParamKey].(map[interface{}]interface{})["title"] = update.Title
	metadata[pageParamKey].(map[interface{}]interface{})["keywords"] = update.Keywords
	metadata[pageParamKey].(map[interface{}]interface{})["description"] = update.Description

	slog.DebugContext(ctx, "font matter", slog.Any("metadata", metadata))

	// Serialize the updated front matter back to YAML
	var buf bytes.Buffer
	buf.WriteString("---\n")

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(metadata); err != nil {
		return err
	}
	buf.WriteString("---\n")

	// Append the rest of the content
	_, err = buf.Write(content)
	if err != nil {
		return fmt.Errorf("failed to append to buffer: %s", err)
	}

	// Write the updated content back to the file
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write to file (%s): %s", path, err)
	}

	return nil
}
