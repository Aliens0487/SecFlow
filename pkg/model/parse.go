package model

import (
	"fmt"
	"log"
	"strings"

	"github.com/Aliens0487/secFlow/pkg/model/bpmn"
)

// BuildGraph constructs a BPMNGraph from Definitions
func BuildGraph(defs *bpmn.Definitions) *BPMNGraph {
	graph := NewBPMNGraph()
	graph.MessageCount = len(defs.Collab.MessageFlow)

	graph.Participants = make([]Participant, len(defs.Collab.Participant))
	log.Println("Participant count:", len(graph.Participants))

	// Add processes and their elements (nodes)
	for i, process := range defs.Process {
		id, participantRaw := findParticpantID(defs.Collab.Participant, process.ID)

		ownerStr := strings.Split(participantRaw.PublicKey, ",")
		var party Participant

		party.PublicKey[0].SetString(ownerStr[0], 10)
		party.PublicKey[1].SetString(ownerStr[1], 10)
		party.ID = id
		party.Name = participantRaw.Name
		graph.Participants[i] = party

		// Add start events
		for _, startEvent := range process.StartEvent {
			graph.addNode(startEvent.ID, "StartEvent", "")
		}

		// Add tasks
		for _, task := range process.Tasks {
			if task.Variables != "" {
				graph.Variables = append(graph.Variables, strings.Split(task.Variables, ",")...)
			}
			if task.Type == "paymentTask" {
				graph.addNodeWithOwner(task.ID, "PaymentTask", task.Name, party)
				partId, _ := findParticpantIdByName(defs.Collab.Participant, task.Participant)
				graph.Nodes[task.ID].Payment = Payment{Receiver: partId, Amount: task.Amount}
			} else {
				graph.addNodeWithOwner(task.ID, "Task", task.Name, party)
			}
		}

		// Add end events
		for _, endEvent := range process.EndEvent {
			graph.addNode(endEvent.ID, "EndEvent", "")
		}

		// Add intermediate events (catch and throw)
		for _, event := range process.IntermediateCatchEvent {
			graph.addNodeWithOwner(event.ID, "IntermediateCatchEvent", "", party)
			message := findMessage(event.ID, defs.Collab.MessageFlow)
			if message != -1 {
				graph.MessageMap[event.ID] = message
			} else {
				panic("Message not found for intermediate catch event: " + event.ID)
			}
		}
		for _, event := range process.IntermediateThrowEvent {
			graph.addNodeWithOwner(event.ID, "IntermediateThrowEvent", "", party)
			message := findMessage(event.ID, defs.Collab.MessageFlow)
			if message != -1 {
				graph.MessageMap[event.ID] = message
			} else {
				panic("Message not found for intermediate throw event: " + event.ID)
			}
		}

		// Add gateways (you could add others like Exclusive, Parallel, etc.)
		// Add exclusive, parallel, and inclusive gateways similarly
		for _, gateway := range process.ExclusiveGateways {
			graph.addNode(gateway.ID, "ExclusiveGateway", "")
		}

		// Add parallel gateways
		for _, gateway := range process.ParallelGateways {
			graph.addNode(gateway.ID, "ParallelGateway", "")
		}

		// Add inclusive gateways
		for _, gateway := range process.InclusiveGateways {
			graph.addNode(gateway.ID, "InclusiveGateway", "")
		}

		// Add sequence flows (edges)
		for _, flow := range process.SequenceFlows {
			graph.addEdge(flow.ID, flow.Name, flow.SourceRef, flow.TargetRef)
		}
	}

	return graph
}

// addNode adds a node to the BPMNGraph
func (g *BPMNGraph) addNode(id, nodeType, name string) {
	if _, exists := g.Nodes[id]; !exists {
		g.Nodes[id] = &Node{
			ID:   id,
			Type: nodeType,
			Name: name,
		}
	}
}

func (g *BPMNGraph) addNodeWithOwner(id, nodeType, name string, owner Participant) {
	if _, exists := g.Nodes[id]; !exists {

		g.Nodes[id] = &Node{
			ID:    id,
			Type:  nodeType,
			Name:  name,
			Owner: owner,
		}
	}
}

// addEdge adds an edge (sequence flow) to the BPMNGraph
func (g *BPMNGraph) addEdge(id, name, sourceID, targetID string) {
	sourceNode, sourceExists := g.Nodes[sourceID]
	targetNode, targetExists := g.Nodes[targetID]

	if sourceExists && targetExists {
		edge := &Edge{
			ID:        id,
			Name:      name,
			SourceRef: sourceNode,
			TargetRef: targetNode,
		}

		// Add the edge to the graph
		g.Edges[id] = edge

		// Connect nodes with this edge
		sourceNode.Outgoing = append(sourceNode.Outgoing, edge)
		targetNode.Incoming = append(targetNode.Incoming, edge)
	} else {
		fmt.Printf("Error: One of the nodes for flow %s not found (source: %s, target: %s)\n", id, sourceID, targetID)
	}
}

func findMessage(nodeId string, messages []bpmn.MessageFlow) int {
	for i, message := range messages {
		if message.SourceRef == nodeId || message.TargetRef == nodeId {
			return i
		}
	}
	return -1
}

func findParticpantID(particpiants []bpmn.Participant, id string) (int, *bpmn.Participant) {
	for i, participant := range particpiants {
		if participant.ProcessRef == id {
			return i, &participant
		}
	}
	return -1, nil
}

func findParticpantIdByName(particpiants []bpmn.Participant, name string) (int, *bpmn.Participant) {
	for i, participant := range particpiants {
		if participant.Name == name {
			return i, &participant
		}
	}
	return -1, nil
}
