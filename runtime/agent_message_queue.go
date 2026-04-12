package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Agent Message Queue — Phase 51
// File-based async communication between agents via brain_v4/_inbox/.
// Format: from_to_timestamp.msg
// Polling interval: 50s. Processed messages move to _archive/.
// ============================================================================

// AgentMessage represents a message between agents
type AgentMessage struct {
	From      string
	To        string
	Timestamp time.Time
	Content   string
	FilePath  string
}

// SendAgentMessage writes a message file to the inbox
func SendAgentMessage(brainRoot, from, to, content string) error {
	inboxDir := filepath.Join(brainRoot, "_inbox", "agent_messages")
	_ = os.MkdirAll(inboxDir, 0755)

	filename := fmt.Sprintf("%s_to_%s_%d.msg", from, to, time.Now().UnixNano())
	msgPath := filepath.Join(inboxDir, filename)

	msgContent := fmt.Sprintf("from: %s\nto: %s\ntime: %s\n---\n%s",
		from, to, time.Now().Format(time.RFC3339), content)

	return os.WriteFile(msgPath, []byte(msgContent), 0644)
}

// ReadAgentMessages reads all messages addressed to the given agent
func ReadAgentMessages(brainRoot, agentName string) ([]AgentMessage, error) {
	inboxDir := filepath.Join(brainRoot, "_inbox", "agent_messages")
	if _, err := os.Stat(inboxDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		return nil, err
	}

	var messages []AgentMessage
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".msg") {
			continue
		}
		// Check if message is addressed to this agent
		if !strings.Contains(e.Name(), "_to_"+agentName+"_") {
			continue
		}

		msgPath := filepath.Join(inboxDir, e.Name())
		data, err := os.ReadFile(msgPath)
		if err != nil {
			continue
		}

		msg := parseAgentMessage(string(data), msgPath)
		messages = append(messages, msg)
	}

	return messages, nil
}

// ArchiveMessage moves a processed message to _archive/
func ArchiveMessage(brainRoot string, msg AgentMessage) error {
	archiveDir := filepath.Join(brainRoot, "_archive", "agent_messages")
	_ = os.MkdirAll(archiveDir, 0755)

	destPath := filepath.Join(archiveDir, filepath.Base(msg.FilePath))
	return os.Rename(msg.FilePath, destPath)
}

// ProcessMessages reads, handles, and archives messages for an agent
func ProcessMessages(brainRoot, agentName string, handler func(AgentMessage)) error {
	messages, err := ReadAgentMessages(brainRoot, agentName)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		handler(msg)
		_ = ArchiveMessage(brainRoot, msg)
	}

	if len(messages) > 0 {
		RecordAudit(brainRoot, "agent_mq", "processed",
			agentName, fmt.Sprintf("%d messages", len(messages)), true)
	}

	return nil
}

// parseAgentMessage parses a .msg file content
func parseAgentMessage(content, filePath string) AgentMessage {
	msg := AgentMessage{FilePath: filePath}

	lines := strings.Split(content, "\n")
	bodyStart := false
	var bodyLines []string

	for _, line := range lines {
		if line == "---" {
			bodyStart = true
			continue
		}
		if bodyStart {
			bodyLines = append(bodyLines, line)
			continue
		}
		if strings.HasPrefix(line, "from: ") {
			msg.From = strings.TrimPrefix(line, "from: ")
		} else if strings.HasPrefix(line, "to: ") {
			msg.To = strings.TrimPrefix(line, "to: ")
		} else if strings.HasPrefix(line, "time: ") {
			t, err := time.Parse(time.RFC3339, strings.TrimPrefix(line, "time: "))
			if err == nil {
				msg.Timestamp = t
			}
		}
	}

	msg.Content = strings.Join(bodyLines, "\n")
	return msg
}
