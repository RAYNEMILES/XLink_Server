package main

import (
	"sync"
)

var LogLevel uint32 = 6
var PlatformID = int32(1)
var LogName = "press_test"

var ReliabilityUserA = 1234567
var ReliabilityUserB = 1234567

var coreMgrLock sync.RWMutex
var allLoginMgr map[int]*CoreNode

var allLoginMgrtmp []*CoreNode

var userLock sync.RWMutex

var allUserID []string
var allToken []string

// var allWs []*interaction.Ws
var sendSuccessCount, sendFailedCount int
var sendSuccessLock sync.RWMutex
var sendFailedLock sync.RWMutex

var msgNumInOneClient = 0

// var Msgwg sync.WaitGroup
var sendMsgClient = 0

var MaxNumGoroutine = 100000

const (
	StepRegister         = 0
	StepRegisterAndLogin = 1
	StepSendMessage      = 2
)
