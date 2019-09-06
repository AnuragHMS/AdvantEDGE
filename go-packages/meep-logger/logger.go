/*
 * Copyright (c) 2019
 * InterDigital Communications, Inc.
 * All rights reserved.
 *
 * The information provided herein is the proprietary and confidential
 * information of InterDigital Communications, Inc.
 */

package logger

import (
	log "github.com/sirupsen/logrus"
)

var componentName string
type Fields map[string]interface{}

func MeepTextLogInit(name string) {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
	componentName = name
}

func MeepJSONLogInit(name string) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	componentName = name
}

func Info(args ...interface{}) {
	log.WithFields(log.Fields{
		"meep.component": componentName,
	}).Info(args...)
}

func Debug(args ...interface{}) {
	log.WithFields(log.Fields{
		"meep.component": componentName,
	}).Debug(args...)
}

func Warn(args ...interface{}) {
	log.WithFields(log.Fields{
		"meep.component": componentName,
	}).Warn(args...)
}
func Error(args ...interface{}) {
	log.WithFields(log.Fields{
		"meep.component": componentName,
	}).Error(args...)
}
func Panic(args ...interface{}) {
	log.WithFields(log.Fields{
		"meep.component": componentName,
	}).Panic(args...)
}

func Fatal(args ...interface{}) {
	log.WithFields(log.Fields{
		"meep.component": componentName,
	}).Fatal(args...)
}

func WithFields(fields Fields) *log.Entry {
	return log.WithFields(log.Fields(fields))
}