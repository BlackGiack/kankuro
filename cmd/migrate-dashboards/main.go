// Command migrate-dashboards migrates a stopped Kankuro configuration offline.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/dashboardmigration"
	"github.com/BlackGiack/kankuro/internal/layout"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "dashboard migration: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) (returnErr error) {
	var configDir string
	var apply, rollback bool
	flags := flag.NewFlagSet("migrate-dashboards", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&configDir, "config-dir", os.Getenv("KANKURO_CONFIG_DIR"), "configuration directory")
	flags.BoolVar(&apply, "apply", false, "stage and publish the migration")
	flags.BoolVar(&rollback, "rollback", false, "restore the backup before the new layout has been used")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || (apply && rollback) {
		return errors.New("unexpected arguments")
	}
	dir, err := config.ResolveDirectory(configDir)
	if err != nil {
		return err
	}
	lock, err := layout.AcquireLock(dir)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, lock.Close()) }()
	if rollback {
		if errAt46 := dashboardmigration.Rollback(dir); errAt46 != nil {
			return errAt46
		}
		_, errAt49 := fmt.Fprintln(output, "Rollback completed; original records restored.")
		return errAt49
	}
	if apply {
		backup, already, errAt53 := dashboardmigration.Apply(dir)
		if errAt53 != nil {
			return errAt53
		}
		if already {
			_, errAt53 = fmt.Fprintln(output, "Dashboard layout is already migrated.")
		} else {
			_, errAt53 = fmt.Fprintf(output, "Migration completed. Backup: %s\n", backup)
		}
		return errAt53
	}
	inventory, err := dashboardmigration.InspectLegacy(dir)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Dry run: %d cards, %d sessions, %d operations, %d messages. No migration was applied.\n",
		inventory.Cards, inventory.Sessions, inventory.Operations, inventory.Messages)
	return err
}
