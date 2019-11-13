package com

type RunStatus string

const (
	OfflineStatus RunStatus = "offline"
	OnlineStatus  RunStatus = "online"
	BreakStatus   RunStatus = "break"
)

type EnvType string

const (
	EnvDev  EnvType = "develop"
	EnvProd EnvType = "product"
	EnvTest EnvType = "test"
)
