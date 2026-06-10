package seed

import (
	"context"
	"log"
)

// RunStartupSeed applies idempotent startup seeding based on install state and SEED_MODE.
func RunStartupSeed(ctx context.Context, installed bool, seedMode string, seeder *Seeder, seedRBAC func(context.Context) error) error {
	if installed {
		switch seedMode {
		case "blank":
			if err := seeder.BlankSiteSeed(ctx); err != nil {
				return err
			}
		case "none":
			log.Println("SEED_MODE=none, skipping seed")
			return nil
		case "demo":
			if err := seeder.DemoSiteSeed(ctx); err != nil {
				return err
			}
		default:
			if err := seeder.DemoSiteSeed(ctx); err != nil {
				return err
			}
		}
		return seedRBAC(ctx)
	}

	switch seedMode {
	case "demo":
		if err := seeder.DemoSiteSeed(ctx); err != nil {
			return err
		}
		return seedRBAC(ctx)
	case "blank":
		if err := seeder.BlankSiteSeed(ctx); err != nil {
			return err
		}
		return seedRBAC(ctx)
	case "none":
		log.Println("Not installed; SEED_MODE=none — complete /setup wizard in the browser")
	default:
		log.Println("Not installed; awaiting /setup wizard (set SEED_MODE=demo for dev auto-seed)")
	}
	return nil
}
