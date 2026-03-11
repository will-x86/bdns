package db

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

type BlocklistEntry struct {
	Name         string
	Url          string
	Category     string
	LastSyncedAt int64
	EntryCount   int
}

// Fakenews, Gambling, Porn, Social, Unified hosts ( adware + malware)
var StevenBlackSources = []BlocklistEntry{
	{
		Name:     "StevenBlack FakeNews",
		Url:      "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-only/hosts",
		Category: "fakenews",
	},
	{
		Name:     "StevenBlack Gambling",
		Url:      "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/gambling-only/hosts",
		Category: "gambling",
	},
	{
		Name:     "StevenBlack Porn",
		Url:      "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/porn-only/hosts",
		Category: "porn",
	},
	{
		Name:     "StevenBlack Social",
		Url:      "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/social-only/hosts",
		Category: "social",
	},
	{
		Name:     "StevenBlack adware+malware",
		Url:      "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/unified-hosts-filtering-only/hosts",
		Category: "unified",
	},
}

func sourceIDForName(name, url, category string) (string, error) {
	var id string
	err := db.QueryRow(`SELECT id FROM blocklist_sources WHERE name = ?`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	err = db.QueryRow(
		`INSERT INTO blocklist_sources (name, url, category, last_synced_at)
		 VALUES (?, ?, ?, unixepoch()) RETURNING id`,
		name, url, category,
	).Scan(&id)
	return id, err
}

func Ingest() error {
	for _, source := range StevenBlackSources {
		if err := ingestSource(source); err != nil {
			return fmt.Errorf("ingest %s: %w", source.Name, err)
		}
	}
	return nil
}
func ingestSource(source BlocklistEntry) error {
	sourceID, err := sourceIDForName(source.Name, source.Url, source.Category)
	if err != nil {
		return fmt.Errorf("sourceIDForName: %w", err)
	}

	resp, err := http.Get(source.Url)

	if err != nil {
		return fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}
	defer tx.Rollback()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		domain := fields[1]

		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO blocklist_entries (source_id, domain,category) VALUES (?, ?, ?)`,
			sourceID, domain, source.Category,
		); err != nil {
			return fmt.Errorf("insert entry: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner.Err: %w", err)
	}

	if _, err := tx.Exec(
		`UPDATE blocklist_sources SET last_synced_at = unixepoch() WHERE id = ?`,
		sourceID,
	); err != nil {
		return fmt.Errorf("update last_synced_at: %w", err)
	}

	return tx.Commit()
}
