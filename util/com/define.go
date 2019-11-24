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

type OpType string

const (
	OpCreate OpType = "create"
	OpUpdate OpType = "update"
	OpDelete OpType = "delete"
)

type ConfigWatchEventType string

const (
	CwNothing      ConfigWatchEventType = "nothing"
	CwConfigChange ConfigWatchEventType = "config_change"
	CwRefreshAll   ConfigWatchEventType = "refresh_all"
)
