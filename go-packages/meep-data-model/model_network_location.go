/*
 * Copyright (c) 2020  InterDigital Communications, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the \"License\");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an \"AS IS\" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * MEEP Model
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package model

// Logical network location object
type NetworkLocation struct {
	// Unique network location ID
	Id string `json:"id,omitempty"`
	// Network location name
	Name string `json:"name,omitempty"`
	// Network location type
	Type_   string                  `json:"type,omitempty"`
	NetChar *NetworkCharacteristics `json:"netChar,omitempty"`
	// **DEPRECATED** As of release 1.5.0, replaced by netChar latency
	TerminalLinkLatency int32 `json:"terminalLinkLatency,omitempty"`
	// **DEPRECATED** As of release 1.5.0, replaced by netChar latencyVariation
	TerminalLinkLatencyVariation int32 `json:"terminalLinkLatencyVariation,omitempty"`
	// **DEPRECATED** As of release 1.5.0, replaced by netChar throughputUl and throughputDl
	TerminalLinkThroughput int32 `json:"terminalLinkThroughput,omitempty"`
	// **DEPRECATED** As of release 1.5.0, replaced by netChar packetLoss
	TerminalLinkPacketLoss float64 `json:"terminalLinkPacketLoss,omitempty"`
	// Key/Value Pair Map (string, string)
	Meta map[string]string `json:"meta,omitempty"`
	// Key/Value Pair Map (string, string)
	UserMeta          map[string]string  `json:"userMeta,omitempty"`
	CellularPoaConfig *CellularPoaConfig `json:"cellularPoaConfig,omitempty"`
	Poa4GConfig       *Poa4GConfig       `json:"poa4GConfig,omitempty"`
	Poa5GConfig       *Poa5GConfig       `json:"poa5GConfig,omitempty"`
	PoaWifiConfig     *PoaWifiConfig     `json:"poaWifiConfig,omitempty"`
	GeoData           *GeoData           `json:"geoData,omitempty"`
	PhysicalLocations []PhysicalLocation `json:"physicalLocations,omitempty"`
}
