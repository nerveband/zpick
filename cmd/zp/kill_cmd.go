package main

func runKill(name string) error {
	b, err := loadBackend(true)
	if err != nil {
		return err
	}
	return b.Kill(name)
}
