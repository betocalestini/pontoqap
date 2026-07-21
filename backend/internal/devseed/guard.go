package devseed

import "fmt"

func (c Config) Guard() error {
	if c.SeedAllow {
		return nil
	}
	switch c.AppEnv {
	case "development", "test":
		return nil
	default:
		return fmt.Errorf("seed bloqueado: defina APP_ENV=development ou SEED_ALLOW=true")
	}
}
