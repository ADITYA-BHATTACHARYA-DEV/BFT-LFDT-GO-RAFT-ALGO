# Consensus Algorithm in GoLang, inspired by Raft - Reliable, Replicated, And Fault-Tolerant (CFT or Crash Fault Tolerance)  

![Raft Visualization Demo](https://github.com/yourusername/raft-visualization/raw/main/demo.gif)


<p align="center">
  <img src="https://github.com/user-attachments/assets/93584dbe-1848-4b6b-98a5-fc99d1f05d36" height="200" width="200" />
</p>







An interactive visualization of the Raft consensus algorithm implemented in Go with terminal UI.

## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
  - [Keyboard Controls](#keyboard-controls)
  - [UI Elements](#ui-elements)
- [Implementation Details](#implementation-details)
  - [Raft Components](#raft-components)
  - [Key Functions](#key-functions)
- [Interactions](#interactions)
  - [Normal Operation](#normal-operation)
  - [Failure Scenarios](#failure-scenarios)
- [License](#license)

## Overview

This project provides an interactive visualization of the Raft consensus algorithm, demonstrating:
- Leader election
- Log replication
- Node failures and recovery
- Network partitions
- Command submission to the cluster

The visualization uses a terminal UI built with `tview` to show real-time status of a 5-node Raft cluster.

## Features

- Real-time visualization of node states (Follower/Candidate/Leader)
- Term number tracking for each node
- Commit index and log length monitoring
- Interactive controls:
  - Toggle node crashes
  - Create network partitions
  - Submit commands to the leader
- Color-coded status display
- Detailed logging of Raft operations


<p align="center">
  <img src="https://github.com/user-attachments/assets/8f05b963-5adb-4c54-95d6-b48e7f775464" height="200" width="200" />
</p>

## Getting Started

### Prerequisites

- Go 1.16 or later
- Terminal with 256-color support (for best experience)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/BFT-LFDT-GO-RAFT-ALGO.git
cd raft-visualization

# Install dependencies
go get github.com/gdamore/tcell/v2
go get github.com/rivo/tview

# Run the visualization
go run main.go

```

## Detailed Analysis of Raft Algorithm Implementation

### Core Raft Components

<p align="center">
  <img src="https://github.com/user-attachments/assets/ad574ba7-30bb-4da4-9261-fa6895bb4cdc" height="400" width="600" />
</p>




#### 1. Node States and Transitions
The code implements all three Raft states with clear transitions:

```bash
type NodeState int
const (
    Follower NodeState = iota  // Initial state
    Candidate                  // Transition during elections
    Leader                     // Active state managing cluster
)

```
##### State Machine Logic:
- All nodes start as Followers
- Followers become Candidates after election timeout
- Candidates become Leaders if they receive majority votes
- Leaders revert to Followers if they discover higher terms

#### 2. Election 



![1 1](https://github.com/user-attachments/assets/9b16b7e4-604e-45cf-a996-d67b1eed35ee)

The election system implements all key Raft requirements:

```bash
func (n *Node) startElection() {
    n.state = Candidate
    n.currentTerm++
    n.votedFor = n.ID
    votes := 1
    
    args := RequestVoteArgs{
        Term:         n.currentTerm,
        CandidateID:  n.ID,
        LastLogIndex: lastLogIndex,
        LastLogTerm:  lastLogTerm,
    }
    // Request votes from peers...
}
```

##### Key Election Features:
- **Randomized Timeouts**: 300-600ms range prevents split votes
- **Term Incrementation:** Ensures monotonic term progression
- **Vote Counting:** Tracks votes received across peers
- **Log Completeness Check:** isLogUpToDate() enforces up-to-date log requirement




#### 3. Log Replication System

The log replication implements the full Raft spec:

```bash
type LogEntry struct {
    Term    int
    Command interface{}
}

func (n *Node) broadcastAppendEntries() {
    // For each peer, send:
    args := AppendEntriesArgs{
        Term:         n.currentTerm,
        LeaderID:     n.ID,
        PrevLogIndex: prevLogIndex,
        PrevLogTerm:  prevLogTerm,
        Entries:      entries,
        LeaderCommit: n.commitIndex,
    }
    // Send to followers...
}
```

##### Replication Guarantees:


- **Log Matching:** Checks PrevLogIndex and PrevLogTerm
- **Consistency:** Overwrites conflicting entries
- **Commit Tracking:** Only advances after majority replication
- **Safety:** Never commits entries from previous terms
