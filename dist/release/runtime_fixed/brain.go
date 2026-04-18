package main

// ?怨ｋ뉴??brain.go ?怨ｋ뉴??// Module: Brain Data Structures + Scanner
//
// PROVIDES: scanBrain, runSubsumption, emitBootstrap, growNeuron, fireNeuron, signalNeuron, rollbackNeuron
//   Neuron, Region, Brain, SubsumptionResult (structs)
//   regionPriority, regionIcons, regionKo (maps)
//   counterRegex, dopamineRegex (regex)
//   scanBrain, runSubsumption, findBrainRoot,
//   getFolderBirthTime, activationBar
//
// CONSUMED BY:
//   ALL files ??core data structures used everywhere
//
// DEPENDS ON:
//   (stdlib only ??foundational module)

import (
        "encoding/json"
        "io/fs"
        "os"
        "path/filepath"
        "regexp"
        "sort"
        "strconv"
        "strings"
        "sync"
        "syscall"
        "time"
)

// Region metadata (SSOT hard-fixed from governance_consts.go)
var regionPriority = map[string]int{
	"brainstem":   0,
	"limbic":      1,
	"hippocampus": 2,
	"sensors":     3,
	"cortex":      4,
	"ego":         5,
	"prefrontal":  6,
	"shared":      7,
}
var regionIcons = map[string]string{
	"brainstem":   "🛡️",
	"limbic":      "💓",
	"hippocampus": "📝",
	"sensors":     "👁️",
	"cortex":      "🧠",
	"ego":         "🎭",
	"prefrontal":  "🎯",
	"shared":      "🔗",
}
var regionKo = map[string]string{
	"brainstem":   "양심/본능",
	"limbic":      "감정 필터",
	"hippocampus": "기록/기억",
	"sensors":     "환경 제약",
	"cortex":      "지식/기술",
	"ego":         "성향/톤",
	"prefrontal":  "목표/계획",
	"shared":      "공유 지식",
}

// ??? Neuron = a folder ???
type Neuron struct {
        Name        string    // folder name
        Path        string    // relative path from region root (e.g. "frontend/css/glass_blur20")
        FullPath    string    // absolute path
        Counter     int       // from N.neuron filename (correction count)      
        Contra      int       // from N.contra filename (inhibition count)      
        Dopamine    int       // from dopamineN.neuron filename (reward count)  
        Intensity   int       // Counter - Contra + Dopamine (net activation)   
        Polarity    float64   // net / total (-1.0=pure inhibition, +1.0=pure excitation)
        HasBomb     bool      // bomb.neuron exists
        HasMemory   bool      // memoryN.neuron exists
        HasGoal     bool      // .goal file exists (todo/objective)
        GoalText    string    // content of .goal file if present
