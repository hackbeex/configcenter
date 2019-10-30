package com

type RunStatus string

const (
	OfflineStatus RunStatus = "offline"
	OnlineStatus  RunStatus = "online"
	BreakStatus   RunStatus = "break"
)
