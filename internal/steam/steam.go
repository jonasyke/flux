package steam

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const ReadyOrNotAppID = 1144200

func GetCurrentBuildID(gameModDir string, appID int) (string, error) {
	manifestPath, err := findManifestPath(gameModDir, appID)
	if err != nil {
		return "", err
	}

	file, err := os.Open(manifestPath)
	if err != nil {
		return "", fmt.Errorf("could not open manifest file: %w", err)
	}
	defer file.Close()

	re := regexp.MustCompile(`"buildid"\s+"(\d+)"`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading manifest file: %w", err)
	}

	return "", fmt.Errorf("buildID not found in %s", manifestPath)
}

func findManifestPath(gameModDir string, appID int) (string, error) {
	curr := filepath.Clean(gameModDir)
	manifestName := fmt.Sprintf("appmanifest_%d.acf", appID)

	for {
		if strings.EqualFold(filepath.Base(curr), "steamapps") {
			manifestPath := filepath.Join(curr, manifestName)

			if info, err := os.Stat(manifestPath); err == nil && !info.IsDir() {
				return manifestPath, nil
			}

			entries, readErr := os.ReadDir(curr)
			if readErr != nil {
				return "", fmt.Errorf("found steamapps but could not read it: %w", readErr)
			}

			for _, entry := range entries {
				if strings.EqualFold(entry.Name(), manifestName) {
					return filepath.Join(curr, entry.Name()), nil
				}
			}

			return "", fmt.Errorf("manifest file '%s' not found inside '%s'", manifestName, curr)
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return "", fmt.Errorf("could not locate 'steamapps' directory from path %s", gameModDir)
}
