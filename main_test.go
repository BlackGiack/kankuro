package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
)

func TestEmbeddedMCPCLISkillDeploysThroughConfig(t *testing.T) {
	if len(embeddedMCPCLISkill) == 0 || !bytes.Contains(embeddedMCPCLISkill, []byte("name: mcp-cli")) {
		t.Fatal("embedded skill is empty or missing its canonical metadata")
	}
	previous, hadPrevious := config.GetBuiltinSkillContent("mcp-cli")
	config.SetBuiltinMCPCLISkillContent(embeddedMCPCLISkill)
	t.Cleanup(func() {
		if hadPrevious {
			config.SetBuiltinMCPCLISkillContent(previous)
		} else {
			config.SetBuiltinMCPCLISkillContent(nil)
		}
	})

	workspace := t.TempDir()
	d := driver.NewGenericCLIDriver()
	skill, ok := config.GetBuiltinSkill(config.BuiltinSkillMCPCLI)
	if !ok {
		t.Fatal("built-in skill metadata is missing")
	}
	if err := d.SetSkills([]config.Skill{skill}, workspace); err != nil {
		t.Fatalf("deploy embedded skill: %v", err)
	}
	deployed, err := os.ReadFile(filepath.Join(workspace, ".agents", "skills", "mcp-cli", "SKILL.md"))
	if err != nil {
		t.Fatalf("read deployed skill: %v", err)
	}
	if !bytes.Equal(deployed, embeddedMCPCLISkill) {
		t.Fatal("deployed skill differs from embedded bytes")
	}
}

func TestEmbeddedKankuroPuppetSkillDeploysThroughConfig(t *testing.T) {
	if len(embeddedKankuroPuppetSkill) == 0 || !bytes.Contains(embeddedKankuroPuppetSkill, []byte("name: kankuro-puppet")) {
		t.Fatal("embedded puppet skill is empty or missing its canonical metadata")
	}
	config.SetBuiltinKankuroPuppetSkillContent(embeddedKankuroPuppetSkill)
	t.Cleanup(func() { config.SetBuiltinKankuroPuppetSkillContent(nil) })
	workspace := t.TempDir()
	skill, ok := config.GetBuiltinSkill(config.BuiltinSkillKankuroPuppet)
	if !ok {
		t.Fatal("built-in puppet skill metadata is missing")
	}
	if err := driver.NewGenericCLIDriver().SetSkills([]config.Skill{skill}, workspace); err != nil {
		t.Fatal(err)
	}
	deployed, err := os.ReadFile(filepath.Join(workspace, ".agents", "skills", "kankuro-puppet", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(deployed, embeddedKankuroPuppetSkill) {
		t.Fatal("deployed puppet skill differs from embedded bytes")
	}
}
