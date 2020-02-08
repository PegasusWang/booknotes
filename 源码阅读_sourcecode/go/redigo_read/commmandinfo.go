package main

import (
	"strings"
)

const (
	connectionWatchState = 1 << iota
	connectionMultiState
	connectionSubscribeState
	connectionMonitorState
)

type commandInfo struct {
	// Set or Clear these states on connection.
	Set, Clear int
}

var commandInfos = map[string]commandInfo{
	"WATCH":      {Set: connectionWatchState},
	"UNWATCH":    {Clear: connectionWatchState},
	"MULTI":      {Set: connectionMultiState},
	"EXEC":       {Clear: connectionWatchState | connectionMultiState},
	"DISCARD":    {Clear: connectionWatchState | connectionMultiState},
	"PSUBSCRIBE": {Set: connectionSubscribeState},
	"SUBSCRIBE":  {Set: connectionSubscribeState},
	"MONITOR":    {Set: connectionMonitorState},
}

func init() {
	for n, ci := range commandInfos {
		commandInfos[strings.ToLower(n)] = ci // 保留一份小写命令的映射
	}
}

func lookupCommandInfo(commandName string) commandInfo {
	if ci, ok := commandInfos[commandName]; ok {
		return ci
	}
	return commandInfos[strings.ToUpper(commandName)]
}
