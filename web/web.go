package web

import "embed"

//go:embed template/* static/*
var Assets embed.FS
