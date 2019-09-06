/*
 * MEEP Mobility Group Manager REST API
 *
 * Copyright (c) 2019 InterDigital Communications, Inc. All rights reserved. The information provided herein is the proprietary and confidential information of InterDigital Communications, Inc.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package server

import (
	"net/http"
)

func CreateMobilityGroup(w http.ResponseWriter, r *http.Request) {
	mgCreateMobilityGroup(w, r)
}

func CreateMobilityGroupApp(w http.ResponseWriter, r *http.Request) {
	mgCreateMobilityGroupApp(w, r)
}

func CreateMobilityGroupUe(w http.ResponseWriter, r *http.Request) {
	mgCreateMobilityGroupUe(w, r)
}

func DeleteMobilityGroup(w http.ResponseWriter, r *http.Request) {
	mgDeleteMobilityGroup(w, r)
}

func DeleteMobilityGroupApp(w http.ResponseWriter, r *http.Request) {
	mgDeleteMobilityGroupApp(w, r)
}

func GetMobilityGroup(w http.ResponseWriter, r *http.Request) {
	mgGetMobilityGroup(w, r)
}

func GetMobilityGroupApp(w http.ResponseWriter, r *http.Request) {
	mgGetMobilityGroupApp(w, r)
}

func GetMobilityGroupAppList(w http.ResponseWriter, r *http.Request) {
	mgGetMobilityGroupAppList(w, r)
}

func GetMobilityGroupList(w http.ResponseWriter, r *http.Request) {
	mgGetMobilityGroupList(w, r)
}

func SetMobilityGroup(w http.ResponseWriter, r *http.Request) {
	mgSetMobilityGroup(w, r)
}

func SetMobilityGroupApp(w http.ResponseWriter, r *http.Request) {
	mgSetMobilityGroupApp(w, r)
}