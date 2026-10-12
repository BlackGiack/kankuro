package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BlackGiack/kankuro/internal/ascii/characters"
)

const (
	maxPuppetLines     = 5
	maxPuppetLineWidth = 11
)

func (r *FileRepository) ListPuppetSources() ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listPuppetSourcesLocked()
}

func (r *FileRepository) listPuppetSourcesLocked() ([]string, error) {
	path := filepath.Join(r.dir, PuppetSourcesFilename)
	f := PuppetSourcesFile{}
	if err := ReadYAML(path, &f); err != nil {
		return nil, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.Paths == nil {
		f.Paths = []string{}
	}
	return f.Paths, nil
}

func (r *FileRepository) SavePuppetSources(paths []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if paths == nil {
		paths = []string{}
	}
	return r.WritePuppetSourcesFile(PuppetSourcesFile{
		SchemaVersion: SchemaVersion,
		Paths:         paths,
	})
}

// UpdatePuppetSources reads the current puppet sources, applies fn,
// persists the result, and re-syncs the puppet registry - all under one
// lock acquisition. See UpdateSkillSources for why this exists and what a
// *SyncFailedError return means.
func (r *FileRepository) UpdatePuppetSources(fn func([]string) ([]string, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.listPuppetSourcesLocked()
	if err != nil {
		return err
	}
	updated, err := fn(current)
	if err != nil {
		return err
	}
	if updated == nil {
		updated = []string{}
	}
	if err = r.WritePuppetSourcesFile(PuppetSourcesFile{SchemaVersion: SchemaVersion, Paths: updated}); err != nil {
		return err
	}
	if err = r.syncPuppetsFromSourcesLocked(); err != nil {
		return &SyncFailedError{Err: err}
	}
	return nil
}

func (r *FileRepository) WritePuppetSourcesFile(f PuppetSourcesFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, PuppetSourcesFilename), f)
}

func (r *FileRepository) LoadPuppetsFile() (PuppetsFile, error) {
	path := filepath.Join(r.dir, PuppetsFilename)
	f := PuppetsFile{}
	if err := ReadYAML(path, &f); err != nil {
		return PuppetsFile{}, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.Puppets == nil {
		f.Puppets = []Puppet{}
	}
	return f, nil
}

func (r *FileRepository) WritePuppetsFile(f PuppetsFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, PuppetsFilename), f)
}

func (r *FileRepository) ListPuppets() ([]Puppet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, err := r.LoadPuppetsFile()
	if err != nil {
		return nil, err
	}

	builtins := builtinPuppets()
	builtinNames := make(map[string]bool, len(builtins))
	for _, b := range builtins {
		builtinNames[b.Name] = true
	}

	result := make([]Puppet, 0, len(builtins)+len(f.Puppets))
	result = append(result, builtins...)
	for _, p := range f.Puppets {
		if !p.Builtin && !builtinNames[p.Name] {
			result = append(result, p)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Builtin != result[j].Builtin {
			return result[i].Builtin
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (r *FileRepository) SyncPuppetsFromSources() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.syncPuppetsFromSourcesLocked()
}

func (r *FileRepository) syncPuppetsFromSourcesLocked() error {
	sources, err := r.listPuppetSourcesLocked()
	if err != nil {
		return err
	}

	var customPuppets []Puppet
	var invalidFiles []string

	for _, src := range sources {
		expanded, expErr := expandPath(src)
		if expErr != nil {
			invalidFiles = append(invalidFiles, fmt.Sprintf("%s: %v", src, expErr))
			continue
		}
		entries, dirErr := os.ReadDir(expanded)
		if dirErr != nil {
			invalidFiles = append(invalidFiles, fmt.Sprintf("%s: %v", src, dirErr))
			continue
		}
		for _, entry := range entries {
			puppet, errStr := readPuppetFile(expanded, entry)
			if errStr != "" {
				invalidFiles = append(invalidFiles, errStr)
				continue
			}
			if puppet != nil {
				customPuppets = append(customPuppets, *puppet)
			}
		}
	}

	return writePuppetsWithDedup(r, customPuppets, invalidFiles)
}

// writePuppetsWithDedup merges custom puppets with builtins, deduplicating
// by name (builtins win), and returns an error describing any skipped files.
func writePuppetsWithDedup(r *FileRepository, customPuppets []Puppet, invalidFiles []string) error {
	builtins := builtinPuppets()
	builtinNames := make(map[string]bool, len(builtins))
	for _, b := range builtins {
		builtinNames[b.Name] = true
	}

	var deduped []Puppet
	customNames := make(map[string]bool)
	for _, p := range customPuppets {
		if builtinNames[p.Name] {
			invalidFiles = append(invalidFiles, fmt.Sprintf("%s: name collides with built-in puppet", p.Name))
			continue
		}
		if customNames[p.Name] {
			invalidFiles = append(invalidFiles, fmt.Sprintf("%s: duplicate name across sources", p.Name))
			continue
		}
		customNames[p.Name] = true
		deduped = append(deduped, p)
	}

	allPuppets := make([]Puppet, 0, len(builtins)+len(deduped))
	allPuppets = append(allPuppets, deduped...)

	if writeErr := r.WritePuppetsFile(PuppetsFile{SchemaVersion: SchemaVersion, Puppets: allPuppets}); writeErr != nil {
		return writeErr
	}

	if len(invalidFiles) > 0 {
		return fmt.Errorf("skipped invalid puppet files: %s", strings.Join(invalidFiles, "; "))
	}
	return nil
}

// readPuppetFile reads a single .puppet file, truncates oversized content,
// and returns a Puppet value. Returns nil Puppet and empty string for hidden
// files, directories, or non-.puppet extensions. Returns non-empty errStr on I/O error.
func readPuppetFile(sourceDir string, entry os.DirEntry) (*Puppet, string) {
	if strings.HasPrefix(entry.Name(), ".") {
		return nil, ""
	}
	if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".puppet") {
		return nil, ""
	}

	fullPath := filepath.Join(sourceDir, entry.Name())
	data, fileErr := os.ReadFile(fullPath)
	if fileErr != nil {
		return nil, fmt.Sprintf("%s: %v", entry.Name(), fileErr)
	}

	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	lines := strings.Split(text, "\n")
	if len(lines) > maxPuppetLines {
		lines = lines[:maxPuppetLines]
	}
	for i, line := range lines {
		if len(line) > maxPuppetLineWidth {
			lines[i] = line[:maxPuppetLineWidth]
		}
	}

	stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
	return &Puppet{
		ID:      puppetIDFromSource(fullPath),
		Name:    stem,
		Source:  fullPath,
		Builtin: false,
		Lines:   lines,
	}, ""
}

func builtinPuppets() []Puppet {
	defs := characters.BuiltinPuppetLines()
	puppets := make([]Puppet, 0, len(defs))
	for _, d := range defs {
		lines := make([]string, len(d.Lines))
		copy(lines, d.Lines)
		puppets = append(puppets, Puppet{
			ID:      fmt.Sprintf("builtin:%s", d.Name),
			Name:    d.Name,
			Builtin: true,
			Lines:   lines,
		})
	}
	return puppets
}

func puppetIDFromSource(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}
