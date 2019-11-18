/*
 * MEEP Model
 *
 * Copyright (c) 2019  InterDigital Communications, Inc Licensed under the Apache License, Version 2.0 (the \"License\"); you may not use this file except in compliance with the License. You may obtain a copy of the License at      http://www.apache.org/licenses/LICENSE-2.0  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an \"AS IS\" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package model

// Network characteristics object
type NetworkCharacteristics struct {
	// Latency in ms
	Latency int32 `json:"latency,omitempty"`
	// Latency variation in ms
	LatencyVariation int32 `json:"latencyVariation,omitempty"`
	// Throughput limit in Mbps
	Throughput int32 `json:"throughput,omitempty"`
	// Packet loss percentage
	PacketLoss float64 `json:"packetLoss,omitempty"`
}
