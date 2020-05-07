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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	dkm "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-data-key-mgr"
	log "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-logger"
	redis "github.com/InterDigitalInc/AdvantEDGE/go-packages/meep-redis"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const notFoundStr string = "na"
const monEngineKey string = "mon-engine:"

var baseKey string = dkm.GetKeyRootGlobal() + monEngineKey

//index in array
const EVENT_POD_ADDED = 0
const EVENT_POD_MODIFIED = 1
const EVENT_POD_DELETED = 2

var pod_event_str = [3]string{"pod added", "pod modified", "pod deleted"}

type MonEngineInfo struct {
	PodName              string
	Namespace            string
	MeepApp              string
	MeepOrigin           string
	MeepScenario         string
	Phase                string
	PodInitialized       string
	PodReady             string
	PodScheduled         string
	PodUnschedulable     string
	PodConditionError    string
	ContainerStatusesMsg string
	NbOkContainers       int
	NbTotalContainers    int
	NbPodRestart         int
	LogicalState         string
	StartTime            string
}

var rc *redis.Connector
var redisDBAddr = "meep-redis-master:6379"
var stopChan = make(chan struct{})

// Init - Mon Engine initialization
func Init() (err error) {

	// Connect to Redis DB
	rc, err = redis.NewConnector(redisDBAddr, 0)
	if err != nil {
		log.Error("Failed connection to Redis: ", err)
		return err
	}
	log.Info("Connected to Mon Engine DB")

	// Empty DB
	_ = rc.DBFlush(baseKey)

	return nil
}

// Run - Mon Engine monitoring thread
func Run() (err error) {

	// Start thread to watch k8s pods
	err = k8sConnect()
	if err != nil {
		log.Error("Failed to watch k8s pods")
		return err
	}

	return nil
}

func Stop() {
	close(stopChan)
}

func connectToAPISvr() (*kubernetes.Clientset, error) {

	// Create the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return clientset, nil
}

func printfMonEngineInfo(monEngineInfo MonEngineInfo, reason int) {

	log.Debug("Monitoring Engine info *** ", pod_event_str[reason], " *** ",
		" pod name : ", monEngineInfo.PodName,
		" namespace : ", monEngineInfo.Namespace,
		" meepApp : ", monEngineInfo.MeepApp,
		" meepOrigin : ", monEngineInfo.MeepOrigin,
		" meepScenario : ", monEngineInfo.MeepScenario,
		" phase : ", monEngineInfo.Phase,
		" podInitialized : ", monEngineInfo.PodInitialized,
		" podUnschedulable : ", monEngineInfo.PodUnschedulable,
		" podScheduled : ", monEngineInfo.PodScheduled,
		" podReady : ", monEngineInfo.PodReady,
		" podConditionError : ", monEngineInfo.PodConditionError,
		" ContainerStatusesMsg : ", monEngineInfo.ContainerStatusesMsg,
		" NbOkContainers : ", monEngineInfo.NbOkContainers,
		" NbTotalContainers : ", monEngineInfo.NbTotalContainers,
		" NbPodRestart : ", monEngineInfo.NbPodRestart,
		" LogicalState : ", monEngineInfo.LogicalState,
		" StartTime : ", monEngineInfo.StartTime)
}

func processEvent(obj interface{}, reason int) {
	if pod, ok := obj.(*v1.Pod); ok {

		var monEngineInfo MonEngineInfo

		if reason != EVENT_POD_DELETED {
			podConditionMsg := ""
			podScheduled := "False"
			podReady := "False"
			podInitialized := "False"
			podUnschedulable := "False"
			nbConditions := len(pod.Status.Conditions)
			for i := 0; i < nbConditions; i++ {
				switch pod.Status.Conditions[i].Type {
				case "PodScheduled":
					podScheduled = string(pod.Status.Conditions[i].Status)
				case "Ready":
					podReady = string(pod.Status.Conditions[i].Status)
					if podReady == "False" {
						podConditionMsg = string(pod.Status.Conditions[i].Message)
					}
				case "Initialized":
					podInitialized = string(pod.Status.Conditions[i].Status)
				case "Unschedulable":
					podUnschedulable = string(pod.Status.Conditions[i].Status)
				}
			}

			nbContainers := len(pod.Status.ContainerStatuses)
			okContainers := 0
			restartCount := 0
			reasonFailureStr := ""
			for i := 0; i < nbContainers; i++ {
				if pod.Status.ContainerStatuses[i].Ready {
					okContainers++
				} else {
					if pod.Status.ContainerStatuses[i].State.Waiting != nil {
						reasonFailureStr = pod.Status.ContainerStatuses[i].State.Waiting.Reason
					} else if pod.Status.ContainerStatuses[i].State.Terminated != nil {
						if reasonFailureStr != "" {
							reasonFailureStr = pod.Status.ContainerStatuses[i].State.Terminated.Reason
						}
					}
				}
				//only update if the value is greater than 0, and we keep it
				if restartCount == 0 {
					restartCount = int(pod.Status.ContainerStatuses[i].RestartCount)
				}
			}

			monEngineInfo.PodInitialized = podInitialized
			monEngineInfo.PodUnschedulable = podUnschedulable
			monEngineInfo.PodScheduled = podScheduled
			monEngineInfo.PodReady = podReady
			monEngineInfo.PodConditionError = podConditionMsg
			monEngineInfo.ContainerStatusesMsg = reasonFailureStr
			monEngineInfo.NbOkContainers = okContainers
			monEngineInfo.NbTotalContainers = nbContainers
			monEngineInfo.NbPodRestart = restartCount
		}

		//common for both the add, update and delete
		monEngineInfo.Phase = string(pod.Status.Phase)
		monEngineInfo.PodName = pod.Name
		monEngineInfo.Namespace = pod.Namespace
		monEngineInfo.MeepApp = pod.Labels["meepApp"]
		monEngineInfo.MeepOrigin = pod.Labels["meepOrigin"]
		monEngineInfo.MeepScenario = pod.Labels["meepScenario"]
		if pod.Labels["meepApp"] != "" {
			monEngineInfo.MeepApp = pod.Labels["meepApp"]
		} else {
			monEngineInfo.MeepApp = notFoundStr
		}
		if pod.Labels["meepOrigin"] != "" {
			monEngineInfo.MeepOrigin = pod.Labels["meepOrigin"]
		} else {
			monEngineInfo.MeepOrigin = notFoundStr
		}
		if pod.Labels["meepScenario"] != "" {
			monEngineInfo.MeepScenario = pod.Labels["meepScenario"]
		} else {
			monEngineInfo.MeepScenario = notFoundStr
		}
		monEngineInfo.LogicalState = monEngineInfo.Phase

		//Phase is Running but might not really be because of some other attributes
		//start of override section of the LogicalState by specific conditions

		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			monEngineInfo.LogicalState = "Terminating"
		} else {
			if monEngineInfo.PodReady != "True" {
				monEngineInfo.LogicalState = "Pending"
			} else {
				if monEngineInfo.NbOkContainers < monEngineInfo.NbTotalContainers {
					monEngineInfo.LogicalState = "Failed"
				}
			}
		}
		//end of override section

		printfMonEngineInfo(monEngineInfo, reason)

		if reason == EVENT_POD_DELETED {
			deleteEntryInDB(monEngineInfo)
		} else {
			addOrUpdateEntryInDB(monEngineInfo)
		}
	}
}

func addOrUpdateEntryInDB(monEngineInfo MonEngineInfo) {
	// Populate rule fields
	fields := make(map[string]interface{})
	fields["name"] = monEngineInfo.PodName
	fields["namespace"] = monEngineInfo.Namespace
	fields["meepApp"] = monEngineInfo.MeepApp
	fields["meepOrigin"] = monEngineInfo.MeepOrigin
	fields["meepScenario"] = monEngineInfo.MeepScenario
	fields["phase"] = monEngineInfo.Phase
	fields["initialised"] = monEngineInfo.PodInitialized
	fields["scheduled"] = monEngineInfo.PodScheduled
	fields["ready"] = monEngineInfo.PodReady
	fields["unschedulable"] = monEngineInfo.PodUnschedulable
	fields["condition-error"] = monEngineInfo.PodConditionError
	fields["nbOkContainers"] = monEngineInfo.NbOkContainers
	fields["nbTotalContainers"] = monEngineInfo.NbTotalContainers
	fields["nbPodRestart"] = monEngineInfo.NbPodRestart
	fields["logicalState"] = monEngineInfo.LogicalState
	fields["startTime"] = monEngineInfo.StartTime

	// Make unique key
	key := baseKey + monEngineInfo.MeepOrigin + ":" + monEngineInfo.Namespace + ":" + monEngineInfo.MeepScenario + ":" + monEngineInfo.MeepApp + ":" + monEngineInfo.PodName

	// Set rule information in DB
	err := rc.SetEntry(key, fields)
	if err != nil {
		log.Error("Entry could not be updated in DB for ", monEngineInfo.MeepApp, ": ", err)
	}
}

func deleteEntryInDB(monEngineInfo MonEngineInfo) {

	// Make unique key
	key := baseKey + monEngineInfo.MeepOrigin + ":" + monEngineInfo.Namespace + ":" + monEngineInfo.MeepScenario + ":" + monEngineInfo.MeepApp + ":" + monEngineInfo.PodName

	// Set rule information in DB
	err := rc.DelEntry(key)
	if err != nil {
		log.Error("Entry could not be deleted in DB for ", monEngineInfo.MeepApp, ": ", err)
	}
}

func k8sConnect() (err error) {

	// Connect to K8s API Server
	clientset, err := connectToAPISvr()
	if err != nil {
		log.Error("Failed to connect with k8s API Server. Error: ", err)
		return err
	}

	meepOrigin := "core"

	// Retrieve pods from k8s api with scenario label
	pods, err := clientset.CoreV1().Pods("").List(
		metav1.ListOptions{LabelSelector: fmt.Sprintf("meepOrigin=%s", meepOrigin)})
	if err != nil {
		log.Error("Failed to retrieve services from k8s API Server. Error: ", err)
		return err
	}

	// Log currently installed core pods
	for _, pod := range pods.Items {
		podName := pod.ObjectMeta.Name
		podPhase := pod.Status.Phase
		log.Debug("podName: ", podName, " podPhase: ", podPhase)
	}

	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())

	// also take a look at NewSharedIndexInformer
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				processEvent(obj, EVENT_POD_ADDED)
			},
			DeleteFunc: func(obj interface{}) {
				processEvent(obj, EVENT_POD_DELETED)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				processEvent(newObj, EVENT_POD_MODIFIED)
			},
		},
	)

	go controller.Run(stopChan)
	return nil
}

// Retrieve POD states
// GET /states
func meGetStates(w http.ResponseWriter, r *http.Request) {
	var allPodsStatus PodsStatus
	var filteredPodsStatus PodsStatus

	// Retrieve query parameters
	query := r.URL.Query()
	queryType := query.Get("type")
	querySandbox := query.Get("sandbox")
	queryLong := query.Get("long")

	// Retrieve pod status information
	var err error
	keyName := baseKey + "*"
	if queryLong == "true" {
		err = rc.ForEachEntry(keyName, getPodDetails, &allPodsStatus)
	} else {
		err = rc.ForEachEntry(keyName, getPodStatesOnly, &allPodsStatus)
	}
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter results based on query parameters
	for _, podStatus := range allPodsStatus.PodStatus {

		// Filter on pod type
		if (podStatus.PodType == notFoundStr) ||
			(queryType == "core" && podStatus.PodType != "core") ||
			(queryType == "scenario" && podStatus.PodType != "scenario") {
			continue
		}

		// Filter on sandbox
		if (querySandbox == "" && podStatus.Sandbox != "default") ||
			(querySandbox != "" && querySandbox != "all" && querySandbox != podStatus.Sandbox) {
			continue
		}

		filteredPodsStatus.PodStatus = append(filteredPodsStatus.PodStatus, podStatus)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Format response
	jsonResponse, err := json.Marshal(filteredPodsStatus)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonResponse))
}

func getPodDetails(key string, fields map[string]string, userData interface{}) error {
	podsStatus := userData.(*PodsStatus)
	var podStatus PodStatus
	podStatus.PodType = fields["meepOrigin"]
	podStatus.Sandbox = fields["namespace"]
	if fields["meepApp"] != notFoundStr {
		podStatus.Name = fields["meepApp"]
	} else {
		podStatus.Name = fields["name"]
	}
	podStatus.Namespace = fields["namespace"]
	podStatus.MeepApp = fields["meepApp"]
	podStatus.MeepOrigin = fields["meepOrigin"]
	podStatus.MeepScenario = fields["meepScenario"]
	podStatus.Phase = fields["phase"]
	podStatus.PodInitialized = fields["initialised"]
	podStatus.PodScheduled = fields["scheduled"]
	podStatus.PodReady = fields["ready"]
	podStatus.PodUnschedulable = fields["unschedulable"]
	podStatus.PodConditionError = fields["condition-error"]
	podStatus.NbOkContainers = fields["nbOkContainers"]
	podStatus.NbTotalContainers = fields["nbTotalContainers"]
	podStatus.NbPodRestart = fields["nbPodRestart"]
	podStatus.LogicalState = fields["logicalState"]
	podStatus.StartTime = fields["startTime"]

	podsStatus.PodStatus = append(podsStatus.PodStatus, podStatus)
	return nil
}

func getPodStatesOnly(key string, fields map[string]string, userData interface{}) error {
	podsStatus := userData.(*PodsStatus)
	var podStatus PodStatus
	podStatus.PodType = fields["meepOrigin"]
	podStatus.Sandbox = fields["namespace"]
	if fields["meepApp"] != notFoundStr {
		podStatus.Name = fields["meepApp"]
	} else {
		podStatus.Name = fields["name"]
	}
	podStatus.LogicalState = fields["logicalState"]
	podsStatus.PodStatus = append(podsStatus.PodStatus, podStatus)
	return nil
}
