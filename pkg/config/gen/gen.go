package main

import (
	cfg "github.com/conductorone/baton-retool/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("retool", cfg.Configuration)
}
