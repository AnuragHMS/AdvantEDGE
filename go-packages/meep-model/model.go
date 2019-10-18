/*
 * Copyright (c) 2019  InterDigital Communications, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	ceModel "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-ctrl-engine-model"
	log "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-logger"
	redis "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-redis"
)

// const activeScenarioEvents = "activeScenarioEvents"
const activeScenarioEvents = "ctrl-engine-active"

// const activeScenarioKey = "activeScenarioKey"
const activeScenarioKey = "ctrl-engine:active"

const DbAddress = "meep-redis-master:6379"

var redisTable = 0

// Model - Implements a Meep Model
type Model struct {
	name          string
	module        string
	Active        bool
	subscribed    bool
	ActiveChannel string
	activeKey     string
	listener      func(string, string)
	rc            *redis.Connector
	scenario      *ceModel.Scenario
	svcMap        []ceModel.NodeServiceMaps
	nodeMap       *NodeMap
}

// NewModel - Create a model object
func NewModel(dbAddr string, module string, name string) (m *Model, err error) {
	if name == "" {
		err = errors.New("Missing name")
		log.Error(err)
		return nil, err
	}
	if module == "" {
		err = errors.New("Missing module")
		log.Error(err)
		return nil, err
	}

	m = new(Model)
	m.name = name
	m.module = module
	m.Active = false
	m.subscribed = false
	m.ActiveChannel = activeScenarioEvents
	m.activeKey = activeScenarioKey
	m.scenario = new(ceModel.Scenario)
	m.nodeMap = NewNodeMap()
	m.parseNodes()
	m.updateSvcMap()

	// Connect to Redis DB
	m.rc, err = redis.NewConnector(dbAddr, redisTable)
	if err != nil {
		log.Error("Model ", m.name, " failed connection to Redis:")
		log.Error(err)
		return nil, err
	}
	log.Debug("[", m.module, "] Model created ", m.name)
	return m, nil
}

// SetScenario - Initialize model from JSON string
func (m *Model) SetScenario(j []byte) (err error) {
	err = json.Unmarshal(j, m.scenario)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	m.parseNodes()
	m.updateSvcMap()
	if m.Active {
		err = m.refresh()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetScenario - Get Scenario JSON string
func (m *Model) GetScenario() (j []byte, err error) {
	j, err = json.Marshal(m.scenario)
	return j, err
}

// Activate - Make scenario the active scenario
func (m *Model) Activate() (err error) {
	jsonScenario, err := json.Marshal(m.scenario)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = m.rc.JSONSetEntry(m.activeKey, ".", string(jsonScenario))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = m.rc.Publish(m.ActiveChannel, "")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	m.Active = true
	return nil
}

// Deactivate - Remove the active scenario
func (m *Model) Deactivate() (err error) {
	if m.Active == true {
		m.Active = false
		err = m.rc.JSONDelEntry(m.activeKey, ".")
		if err != nil {
			log.Error(err.Error())
			return err
		}
		err = m.rc.Publish(m.ActiveChannel, "")
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

//Listen - Listen to scenario update events
func (m *Model) Listen(handler func(string, string)) (err error) {
	if handler == nil {
		return errors.New("Nil event handler")
	}
	if !m.subscribed {
		// Subscribe to Pub-Sub events for MEEP Controller
		err = m.rc.Subscribe(m.ActiveChannel)
		if err != nil {
			log.Error("Failed to subscribe to Pub/Sub events. Error: ", err)
			return err
		}
		log.Info("Subscribed to active scenario events (Redis)")
		m.subscribed = true

		m.listener = handler

		// Listen for events
		go func() {
			_ = m.rc.Listen(m.internalListener)
		}()

		// Generate first event to initialize
		m.internalListener(m.ActiveChannel, "")
	}
	return nil
}

// MoveNode - Move a specific UE in the scenario
func (m *Model) MoveNode(nodeName string, destName string) (oldLocName string, newLocName string, err error) {
	moveNode := m.nodeMap.FindByName(nodeName)
	// fmt.Printf("+++ ueNode: %+v\n", ueNode)
	if moveNode == nil {
		return "", "", errors.New("Mobility: " + nodeName + " not found")
	}

	if moveNode.nodeType == "EDGE-APP" {
		oldLocName, newLocName, err = m.moveProc(moveNode, destName)
		if err != nil {
			return "", "", err
		}
	} else {
		oldLocName, newLocName, err = m.movePL(moveNode, destName)
		if err != nil {
			return "", "", err
		}
	}

	err = m.refresh()
	if err != nil {
		return "", "", err
	}
	return oldLocName, newLocName, nil
}

// GetServiceMaps - Extracts the model service maps
func (m *Model) GetServiceMaps() *[]ceModel.NodeServiceMaps {
	return &m.svcMap
}

//UpdateNetChar - Update network characteristics for a node
func (m *Model) UpdateNetChar(nc *ceModel.EventNetworkCharacteristicsUpdate) (err error) {
	updated := false

	ncName := nc.ElementName
	ncType := strings.ToUpper(nc.ElementType)

	// Find the element
	if ncType == "SCENARIO" {
		m.scenario.Deployment.InterDomainLatency = nc.Latency
		m.scenario.Deployment.InterDomainLatencyVariation = nc.LatencyVariation
		m.scenario.Deployment.InterDomainThroughput = nc.Throughput
		m.scenario.Deployment.InterDomainPacketLoss = nc.PacketLoss
		updated = true
	} else {
		n := m.nodeMap.FindByName(ncName)
		// fmt.Printf("+++ node: %+v\n", n)
		if n == nil {
			return errors.New("Did not find " + ncName + " in scenario " + m.name)
		}
		if ncType == "OPERATOR" {
			domain := n.object.(*ceModel.Domain)
			domain.InterZoneLatency = nc.Latency
			domain.InterZoneLatencyVariation = nc.LatencyVariation
			domain.InterZoneThroughput = nc.Throughput
			domain.InterZonePacketLoss = nc.PacketLoss
			updated = true
		} else if ncType == "ZONE-INTER-EDGE" {
			zone := n.object.(*ceModel.Zone)
			zone.InterEdgeLatency = nc.Latency
			zone.InterEdgeLatencyVariation = nc.LatencyVariation
			zone.InterEdgeThroughput = nc.Throughput
			zone.InterEdgePacketLoss = nc.PacketLoss
			updated = true
		} else if ncType == "ZONE-INTER-FOG" {
			zone := n.object.(*ceModel.Zone)
			zone.InterFogLatency = nc.Latency
			zone.InterFogLatencyVariation = nc.LatencyVariation
			zone.InterFogThroughput = nc.Throughput
			zone.InterFogPacketLoss = nc.PacketLoss
			updated = true
		} else if ncType == "ZONE-EDGE-FOG" {
			zone := n.object.(*ceModel.Zone)
			zone.EdgeFogLatency = nc.Latency
			zone.EdgeFogLatencyVariation = nc.LatencyVariation
			zone.EdgeFogThroughput = nc.Throughput
			zone.EdgeFogPacketLoss = nc.PacketLoss
			updated = true
		} else if ncType == "POA" {
			nl := n.object.(*ceModel.NetworkLocation)
			nl.TerminalLinkLatency = nc.Latency
			nl.TerminalLinkLatencyVariation = nc.LatencyVariation
			nl.TerminalLinkThroughput = nc.Throughput
			nl.TerminalLinkPacketLoss = nc.PacketLoss
			updated = true
		} else if ncType == "DISTANT CLOUD" || ncType == "EDGE" || ncType == "FOG" || ncType == "UE" {
			pl := n.object.(*ceModel.PhysicalLocation)
			pl.LinkLatency = nc.Latency
			pl.LinkLatencyVariation = nc.LatencyVariation
			pl.LinkThroughput = nc.Throughput
			pl.LinkPacketLoss = nc.PacketLoss
			updated = true
		} else if ncType == "CLOUD APPLICATION" || ncType == "EDGE APPLICATION" || ncType == "UE APPLICATION" {
			proc := n.object.(*ceModel.Process)
			proc.AppLatency = nc.Latency
			proc.AppLatencyVariation = nc.LatencyVariation
			proc.AppThroughput = nc.Throughput
			proc.AppPacketLoss = nc.PacketLoss
			updated = true
		}

	}
	if updated {
		err = m.refresh()
		if err != nil {
			return err
		}
	}
	return nil
}

//GetScenarioName - Get the scenario name
func (m *Model) GetScenarioName() string {
	fmt.Printf("%+v", m)
	if m.scenario != nil {
		return m.scenario.Name
	}
	return ""
}

//GetNodeNames - Get the list of nodes of a certain type; "" or "ANY" returns all
func (m *Model) GetNodeNames(typ string) []string {
	var l int
	var nm map[string]*Node
	if typ == "" || typ == "ANY" {
		nm = m.nodeMap.nameMap
		l = len(nm)
	} else {
		nm = m.nodeMap.typeMap[typ]
		l = len(nm)
	}
	list := make([]string, 0, l)
	for k := range nm {
		list = append(list, k)
	}
	return list
}

//GetNodeEdges - Get a map of node edges for the current scenario
func (m *Model) GetEdges() (edgeMap map[string]string) {
	edgeMap = make(map[string]string)
	for k, node := range m.nodeMap.nameMap {
		p := reflect.ValueOf(node.parent)
		pName := reflect.Indirect(p).FieldByName("Name")
		if pName.IsValid() {
			edgeMap[k] = pName.String()
			// fmt.Printf("%s (%T) \t\t %s(%T)\n", k, node.object, pName, node.parent)
		}
	}
	return edgeMap
}

// GetNode - Get a node by its name
// 		Returned value is of type interface{}
//    Good practice: returned node should be type asserted with val,ok := node.(someType) to prevent panic
func (m *Model) GetNode(name string) (node interface{}) {
	node = nil
	n := m.nodeMap.nameMap[name]
	if n != nil {
		node = n.object
	}
	return node
}

//---Internal Funcs---

func (m *Model) parseNodes() (err error) {
	if m.scenario.Deployment != nil {
		if m.scenario.Deployment != nil {
			// Parse through scenario and fill external node service mappings
			for iDomain := range m.scenario.Deployment.Domains {
				domain := &m.scenario.Deployment.Domains[iDomain]
				m.nodeMap.AddNode(NewNode(domain.Name, domain.Type_, domain, &domain.Zones, m.scenario.Deployment))
				for iZone := range domain.Zones {
					zone := &domain.Zones[iZone]
					m.nodeMap.AddNode(NewNode(zone.Name, zone.Type_, zone, &zone.NetworkLocations, domain))
					for iNL := range zone.NetworkLocations {
						nl := &zone.NetworkLocations[iNL]
						m.nodeMap.AddNode(NewNode(nl.Name, nl.Type_, nl, &nl.PhysicalLocations, zone))
						for iPL := range nl.PhysicalLocations {
							pl := &nl.PhysicalLocations[iPL]
							m.nodeMap.AddNode(NewNode(pl.Name, pl.Type_, pl, &pl.Processes, nl))
							for iProc := range pl.Processes {
								proc := &pl.Processes[iProc]
								m.nodeMap.AddNode(NewNode(proc.Name, proc.Type_, proc, nil, pl))
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (m *Model) updateSvcMap() (err error) {
	if m.scenario.Deployment == nil {
		m.svcMap = nil
	} else {
		// Parse through scenario and fill external node service mappings
		for _, domain := range m.scenario.Deployment.Domains {
			for _, zone := range domain.Zones {
				for _, nl := range zone.NetworkLocations {
					for _, pl := range nl.PhysicalLocations {
						for _, proc := range pl.Processes {
							if proc.IsExternal {
								// Create new node service map
								var nodeServiceMaps ceModel.NodeServiceMaps
								nodeServiceMaps.Node = proc.Name
								nodeServiceMaps.IngressServiceMap = append(nodeServiceMaps.IngressServiceMap,
									proc.ExternalConfig.IngressServiceMap...)
								nodeServiceMaps.EgressServiceMap = append(nodeServiceMaps.EgressServiceMap,
									proc.ExternalConfig.EgressServiceMap...)

								// Add new map to list
								m.svcMap = append(m.svcMap, nodeServiceMaps)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (m *Model) refresh() (err error) {
	if m.Active == true {
		err = m.rc.JSONDelEntry(m.activeKey, ".")
		if err != nil {
			log.Error(err.Error())
			return err
		}
		jsonScenario, err := json.Marshal(m.scenario)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		err = m.rc.JSONSetEntry(m.activeKey, ".", string(jsonScenario))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		err = m.rc.Publish(m.ActiveChannel, "")
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (m *Model) movePL(node *Node, destName string) (oldLocName string, newLocName string, err error) {
	var pl *ceModel.PhysicalLocation
	var oldNL *ceModel.NetworkLocation
	var newNL *ceModel.NetworkLocation

	// Node is a UE
	pl = node.object.(*ceModel.PhysicalLocation)
	// fmt.Printf("+++ pl: %+v\n", pl)

	oldNL = node.parent.(*ceModel.NetworkLocation)
	// fmt.Printf("+++ oldNL: %+v\n", oldNL)
	if oldNL == nil {
		return "", "", errors.New("MoveNode: " + node.name + " old location not found)")
	}

	newNLNode := m.nodeMap.FindByName(destName)
	// fmt.Printf("+++ newNLNode: %+v\n", newNLNode)
	if newNLNode == nil {
		return "", "", errors.New("MoveNode: " + destName + " not found")
	}
	newNL = newNLNode.object.(*ceModel.NetworkLocation)
	// fmt.Printf("+++ newNL: %+v\n", newNL)

	// Update location if necessary
	if pl != nil && oldNL != nil && newNL != nil && oldNL != newNL {
		log.Debug("Found PL & destination. Updating PL location.")

		// Add PL to new location
		newNL.PhysicalLocations = append(newNL.PhysicalLocations, *pl)
		node.parent = newNL

		var idx int
		for i, x := range oldNL.PhysicalLocations {
			if x.Type_ == "UE" && x.Name == node.name {
				idx = i
				break
			}
		}
		// Remove UE from old location
		//overwrite
		oldNL.PhysicalLocations[idx] = oldNL.PhysicalLocations[len(oldNL.PhysicalLocations)-1]
		//truncate
		oldNL.PhysicalLocations = oldNL.PhysicalLocations[:len(oldNL.PhysicalLocations)-1]
	}

	return oldNL.Name, newNL.Name, nil
}

func (m *Model) moveProc(node *Node, destName string) (oldLocName string, newLocName string, err error) {
	var proc *ceModel.Process
	var oldPL *ceModel.PhysicalLocation
	var newPL *ceModel.PhysicalLocation

	// Node is a process
	proc = node.object.(*ceModel.Process)
	// fmt.Printf("+++ process: %+v\n", proc)
	//process part of a mobility group can't be moved
	if proc.ServiceConfig != nil {
		if proc.ServiceConfig.MeSvcName != "" {
			return "", "", errors.New("Process part of a mobility group cannot be moved ")
		}
	}

	oldPL = node.parent.(*ceModel.PhysicalLocation)
	// fmt.Printf("+++ oldPL: %+v\n", oldPL)
	if oldPL == nil {
		return "", "", errors.New("MoveNode: " + node.name + " old location not found)")
	}

	newPLNode := m.nodeMap.FindByName(destName)
	// fmt.Printf("+++ newPLNode: %+v\n", newPLNode)
	if newPLNode == nil {
		return "", "", errors.New("MoveNode: " + destName + " not found")
	}
	newPL = newPLNode.object.(*ceModel.PhysicalLocation)
	// fmt.Printf("+++ newNL: %+v\n", newNL)

	// Update location if necessary
	if proc != nil && oldPL != nil && newPL != nil && oldPL != newPL {
		log.Debug("Found Process & destination. Updating PL location.")

		// Add PL to new location
		newPL.Processes = append(newPL.Processes, *proc)
		node.parent = newPL

		var idx int
		for i, x := range oldPL.Processes {
			if x.Name == node.name {
				idx = i
				break
			}
		}
		// Remove UE from old location
		//overwrite
		oldPL.Processes[idx] = oldPL.Processes[len(oldPL.Processes)-1]
		//truncate
		oldPL.Processes = oldPL.Processes[:len(oldPL.Processes)-1]
	}

	return oldPL.Name, newPL.Name, nil
}

func (m *Model) internalListener(channel string, payload string) {
	// An update was received - Update the object state and call the external Handler
	// Retrieve active scenario from DB
	j, err := m.rc.JSONGetEntry(m.activeKey, ".")
	log.Debug("Scenario Event:", j)
	if err != nil {
		log.Debug("Scenario was deleted, create a new one")
		// Scenario was deleted
		m.scenario = new(ceModel.Scenario)
		m.nodeMap = NewNodeMap()
		m.parseNodes()
		m.updateSvcMap()
	} else {
		m.SetScenario([]byte(j))
	}

	// external listener
	m.listener(channel, payload)
}
