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
- [Detailed Analysis of Raft Algorithm Implementation](#detailed-analysis-of-raft-algorithm-implementation)
  - [Core Raft Components](#core-raft-components)
  - [Leader Election Phase](#leader-election-phase)

- [Safety Implementation](#safety-properties-implementation)
- [Core Logic & Processes](#core-logic-&-processes)
 
-  [Working Sample](#working-sample)
  - [ScreenShots](#screenshots)
  - [Live Demo Video](#live-demo-video)


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
cd BFT-LFDT-GO-RAFT-ALGO

# Install dependencies
go get github.com/gdamore/tcell/v2
go get github.com/rivo/tview

# Run the visualization
go run main.go

```

## Detailed Analysis of Raft Algorithm Implementation

### Core Raft Components

<p align="center">
  <img src="https://github.com/user-attachments/assets/ad574ba7-30bb-4da4-9261-fa6895bb4cdc" height="600" width="700" />
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

  </br>
  </br>



<p align="center">
  <img src="https://github.com/user-attachments/assets/6f5a56b3-0cb1-4e32-9e06-4d6f9d1534bf" height="3800" width="400" />
</p>


  



## Leader Election Phase
### 1.Follower Timeout:

- Each node runs resetElectionTimer() with random duration

- On timeout, transitions to Candidate via startElection()

### 2.Vote Request:

- Candidate increments term and votes for itself

- Sends **_RequestVote_** RPCs to all peers

- Includes last log index/term for completeness check

### 3.Vote Collection:

- Peers validate candidate's log is up-to-date

- Grant vote if haven't voted this term

- Candidate becomes Leader on majority votes

### Normal Operation (Leader Active)
#### 1.Heartbeat Mechanism:

- Leader runs startHeartbeats() with 50ms interval

- Empty AppendEntries as heartbeat

- Maintains authority and detects failures

#### 2.Command Processing:

```bash
func (n *Node) SubmitCommand(cmd interface{}) {
    n.log = append(n.log, LogEntry{
        Term:    n.currentTerm,
        Command: cmd,
    })
    n.broadcastAppendEntries()
}
```
- Client commands appended to leader's log
- Immediately replicated to followers
- Committed after majority acknowledgement

#### 3.Commit Propagation:
- Leader tracks **nextIndex** and **matchIndex**per follower

- Updates **commitIndex** via **updateCommitIndex()**

- Followers apply committed entries through **applyLogs()**

### Failure Handling

#### 1.Leader Crash Detection

- Followers timeout waiting for heartbeats

- Transition to Candidate and start new election

- Higher term prevents split-brain scenarios

</br>
</br>


<p align="center">
  <img src="https://github.com/user-attachments/assets/f4e495f6-558d-445d-98ed-36bc12ae91ad" height="600" width="600" />
</p>



</br>

### 2.Network Partitions:


```bash
nodes[a].networkPartitioned[b] = true
nodes[b].networkPartitioned[a] = true
```
- Partitioned leader cannot commit entries

- Partition with majority elects new leader

- Recovered partitions reconcile via RPC term checks

### 3.Log Reconciliation:

- Followers truncate logs on inconsistency

- Leader decrements nextIndex on failure

- Eventually finds matching log point


## Safety Properties Implementation

### Election Safety

```bash

func (n *Node) HandleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
    if args.Term > n.currentTerm {
        n.stepDown(args.Term)
    }
    // Additional checks...
}
```
- At most one leader per term (via term comparisons)

- Only up-to-date nodes can become leaders (isLogUpToDate() check)

### Log Safety

```bash
func (n *Node) updateCommitIndex() {
    for N := len(n.log) - 1; N > n.commitIndex; N-- {
        if n.log[N].Term != n.currentTerm {
            continue // Only commit current term entries
        }
        // Majority check...
    }
}

```
- Never overwrite committed entries

- Only leader's current term entries committed directly

- Previous term entries committed indirectly

### State Machine Safely

```bash

func (n *Node) applyLogs() {
    for n.lastApplied < n.commitIndex {
        n.lastApplied++
        entry := n.log[n.lastApplied]
        n.applyCh <- ApplyMsg{
            CommandValid: true,
            Command:      entry.Command,
            CommandIndex: n.lastApplied,
        }
    }
}
```

- Applies entries in order

- Only applies committed entries

- Idempotent application (can survive crashes)

</br>
</br>

![server-states-l](https://github.com/user-attachments/assets/5cba9ef4-f59a-44c2-842c-3c1f1d493f02)



</br>

## Optimization Details

- **Batching:** The implementation could batch multiple log entries in single AppendEntries RPCs but currently sends entries individually.

- **Pipelining:** The code implicitly pipelines requests by not waiting for previous RPCs to complete before sending new ones.

- **Transport Efficiency:** Heartbeats are empty AppendEntries RPCs, minimizing bandwidth when idle.





</br>
</br>

## Core Logic & Processes
### 🧠 Consensus Mechanism (Log Replication)
**Goal:** Ensure that all nodes agree on the same sequence of commands (log entries).

**How it works in your code:**
Client submits command:



```bash
node.SubmitCommand(command)
```
- Only the leader accepts commands (if n.state != Leader || n.crashed { return }).

- Appends a LogEntry to the leader's local log:


```bash
n.log = append(n.log, LogEntry{Term: n.currentTerm, Command: cmd})
```
#### Leader broadcasts to followers:

```bash
n.broadcastAppendEntries()
```
#### Sends AppendEntriesArgs to all peers:


``` bash
entries := n.log[nextIndex:]
```
#### Followers validate and append:


```bash
func (n *Node) HandleAppendEntries(...)
```
- If log doesn't match at PrevLogIndex, reject (enforcing consistency).

- If terms match, append new entries:


```bash
n.log = append(n.log, entry)
```
- Commit & Apply:

- Once the leader knows a log entry is replicated on a majority, it commits:


```bash
n.updateCommitIndex()
```
- Then, each node applies the entry:

```bash
n.applyLogs()
```
### 📌 This mechanism ensures that all non-faulty nodes eventually agree on the same ordered list of commands.

## 🔐 Fault Tolerance: Crashes & Message Drops
### Goal: System keeps working despite node crashes or communication issues.

- **Crashes**
- Nodes can be crashed manually with keys 0–4:

```bash
node.crashed = !node.crashed
```
- Crashed nodes:

- Do not process elections, RPCs, or timers.

- Skip logic in HandleRequestVote, HandleAppendEntries, etc.:


```bash
if n.crashed {
    return
}
```
- Recovering from crash
- When revived (!node.crashed), the node resumes normal operation:


```bash
node.resetElectionTimer()
if node.state == Leader {
    node.startHeartbeats()
}
```
- Message drops
- Simulated indirectly via network partitions (below).

## 🔌 Network Partitions, Retries, and Quorum Logic
### Network Partitions
### Goal: Test how the system behaves when communication between some nodes is cut.

- You simulate partitions using keyboard input: pXY

```bash
nodes[a].networkPartitioned[b] = true
```
- This marks nodes as unable to communicate:

- **In startElection:**

```bash
if n.networkPartitioned[peer.ID] { continue }
```
- **In broadcastAppendEntries:**

```bash
if n.networkPartitioned[p.ID] { return }
```
### Effect:

- Partitioned nodes can't vote or replicate logs.

- Leader may lose quorum and be unable to commit new commands.

## Raft ensures safety (no split-brain), even during partitions.

### Retries (in Log Replication)
- If a follower rejects AppendEntries (log mismatch):

```bash
if !reply.Success {
    if n.nextIndex[p.ID] > 0 {
        n.nextIndex[p.ID]-- // Decrement nextIndex and retry
    }
}
```
- This retry loop ensures the leader backs up to the matching log index and reattempts replication until success.

## Quorum Logic
### Quorum = majority(n):

```bash
func majority(n int) int {
    return (n / 2) + 1
}
```
### Used in:

- Elections: Candidate wins if it gets a majority of votes:


```bash
if votes >= majority(len(n.peers)+1) { ... }
```
### Commitment: Leader only commits if a log entry is replicated on a majority:


```bash
if count >= majority(len(n.peers)+1) { ... }
```
### This ensures:

- Fault tolerance (can handle up to ⌊n/2⌋ failures).

- Linearizability (all committed entries are consistent).

## 📺 Output: Leader Changes, Agreement Steps, State Updates
### Leader Changes
#### Tracked via:
```bash
n.state == Leader transition in becomeLeader():
```


```bash
n.state = Leader
fmt.Printf("Node %d became leader in term %d\n", n.ID, n.currentTerm)
```

### Agreement Steps (Log Commit)
```bash
In applyLogs():
```

```bash
fmt.Printf("Node %d applied command: %v\n", n.ID, entry.Command)
```
### Shows which commands were committed and by which node.

## UI Updates
- Every 300ms, the UI updates node state:


```bash
table.SetCell(..., state)
```
## Shows:


- Node state (Follower/Candidate/Leader)

- Term

- Commit index

- Log length

- Leader ID

- Crash status (Up/Down)

- Partitioned peers


</br>


### User Input Controls
- **Crash node:** keys 0–4

- **Partition nodes:** pXY (e.g., p13)

- **Submit command to leader:** type and press Enter

## Working Sample

### ScreenShots 

![Screenshot 2025-05-26 180138](https://github.com/user-attachments/assets/cfd8547d-0101-4abd-96b3-6c6e8e7d9502)

</br>
</br>


![Screenshot 2025-05-26 175927](https://github.com/user-attachments/assets/824883f7-c3d9-4a0c-8893-67861efeb4e3)


</br>
</br>

### Live Demo Video:




https://github.com/user-attachments/assets/8cc2d881-c795-4aae-b95d-a4f8c82d159c


