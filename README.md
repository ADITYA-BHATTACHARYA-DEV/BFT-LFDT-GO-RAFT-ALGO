# Raft Consensus Algorithm Visualization

![Raft Visualization Demo](https://github.com/yourusername/raft-visualization/raw/main/demo.gif)

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

## Getting Started

### Prerequisites

- Go 1.16 or later
- Terminal with 256-color support (for best experience)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/raft-visualization.git
cd raft-visualization

# Install dependencies
go get github.com/gdamore/tcell/v2
go get github.com/rivo/tview

# Run the visualization
go run main.go
