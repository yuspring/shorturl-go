package web

import "embed"

// Assets 包含所有的前端資源 (模板與靜態檔案)
//
//go:embed template/* static/*
var Assets embed.FS
