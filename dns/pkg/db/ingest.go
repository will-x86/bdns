package db

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

const (
	stevenBlackSocialURL  = "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/social-only/hosts"
	stevenBlackSocialName = "StevenBlack-social"
)

func sourceIDForName(name, url, category string) (string, error) {
	var id string
	err := db.QueryRow(`SELECT id FROM blocklist_sources WHERE name = ?`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	err = db.QueryRow(
		`INSERT INTO blocklist_sources (name, url, category, enabled, last_synced_at)
		 VALUES (?, ?, ?, 1, unixepoch()) RETURNING id`,
		name, url, category,
	).Scan(&id)
	return id, err
}

func IngestStevenBlack() error {
	resp, err := http.Get(stevenBlackSocialURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	sourceID, err := sourceIDForName(stevenBlackSocialName, stevenBlackSocialURL, "social")
	if err != nil {
		return fmt.Errorf("resolve source: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO blocklist_entries (domain, source_id, category) VALUES (?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	var category string
	var total int

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if payload, ok := strings.CutPrefix(line, "# "); ok {
			if !strings.Contains(payload, ":") && !strings.Contains(payload, " ") {
				category = strings.ToLower(payload)
			}
			continue
		}
		if category == "" || !strings.HasPrefix(line, "0.0.0.0 ") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		domain := fields[1]
		if domain == "0.0.0.0" || strings.HasPrefix(domain, "localhost") {
			continue
		}

		if _, err := stmt.Exec(domain, sourceID, category); err != nil {
			return fmt.Errorf("insert %s: %w", domain, err)
		}
		total++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	_, err = db.Exec(
		`UPDATE blocklist_sources SET entry_count = ?, last_synced_at = unixepoch() WHERE id = ?`,
		total, sourceID,
	)
	return err
}
