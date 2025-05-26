package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

type LogEntry struct {
	Term    int
	Command interface{}
}

type Node struct {
	mu                 sync.Mutex
	ID                 int
	peers              []*Node
	currentTerm        int
	votedFor           int
	log                []LogEntry
	commitIndex        int
	lastApplied        int
	state              NodeState
	leaderID           int
	nextIndex          map[int]int
	matchIndex         map[int]int
	electionTimer      *time.Timer
	heartbeatTimer     *time.Timer
	applyCh            chan ApplyMsg
	stopCh             chan struct{}
	crashed            bool
	networkPartitioned map[int]bool
}

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type RequestVoteArgs struct {
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func NewNode(id int) *Node {
	return &Node{
		ID:                 id,
		currentTerm:        0,
		votedFor:           -1,
		state:              Follower,
		applyCh:            make(chan ApplyMsg, 100),
		stopCh:             make(chan struct{}),
		networkPartitioned: make(map[int]bool),
	}
}

func (n *Node) Start() {
	n.resetElectionTimer()
	go n.run()
}

func (n *Node) run() {
	for {
		select {
		case <-n.stopCh:
			return
		}
	}
}

func (n *Node) resetElectionTimer() {
	if n.crashed {
		return
	}
	if n.electionTimer != nil {
		n.electionTimer.Stop()
	}
	timeout := time.Duration(300+rand.Intn(300)) * time.Millisecond
	n.electionTimer = time.AfterFunc(timeout, func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		if n.state == Leader || n.crashed {
			return
		}
		n.startElection()
	})
}

func (n *Node) startElection() {
	if n.crashed {
		return
	}
	n.state = Candidate
	n.currentTerm++
	n.votedFor = n.ID
	votes := 1

	lastLogIndex, lastLogTerm := n.lastLogInfo()
	args := RequestVoteArgs{
		Term:         n.currentTerm,
		CandidateID:  n.ID,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}

	for _, peer := range n.peers {
		if n.networkPartitioned[peer.ID] {
			continue
		}
		go func(p *Node) {
			if p.crashed {
				return
			}
			var reply RequestVoteReply
			p.HandleRequestVote(&args, &reply)

			n.mu.Lock()
			defer n.mu.Unlock()

			if reply.Term > n.currentTerm {
				n.stepDown(reply.Term)
				return
			}

			if reply.VoteGranted {
				votes++
				if votes >= majority(len(n.peers)+1) {
					n.becomeLeader()
				}
			}
		}(peer)
	}
}

func (n *Node) HandleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.crashed {
		return
	}

	reply.Term = n.currentTerm
	reply.VoteGranted = false

	if args.Term < n.currentTerm {
		return
	}

	if args.Term > n.currentTerm {
		n.stepDown(args.Term)
	}

	if (n.votedFor == -1 || n.votedFor == args.CandidateID) &&
		n.isLogUpToDate(args.LastLogIndex, args.LastLogTerm) {
		reply.VoteGranted = true
		n.votedFor = args.CandidateID
		n.resetElectionTimer()
	}
}

func (n *Node) HandleAppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.crashed {
		return
	}

	reply.Term = n.currentTerm
	reply.Success = false

	if args.Term < n.currentTerm {
		return
	}

	n.resetElectionTimer()

	if args.Term > n.currentTerm {
		n.stepDown(args.Term)
	}

	n.state = Follower
	n.leaderID = args.LeaderID

	if args.PrevLogIndex >= 0 &&
		(args.PrevLogIndex >= len(n.log) || n.log[args.PrevLogIndex].Term != args.PrevLogTerm) {
		return
	}

	for i, entry := range args.Entries {
		index := args.PrevLogIndex + 1 + i
		if index < len(n.log) && n.log[index].Term != entry.Term {
			n.log = n.log[:index]
		}
		if index >= len(n.log) {
			n.log = append(n.log, entry)
		}
	}

	if args.LeaderCommit > n.commitIndex {
		n.commitIndex = min(args.LeaderCommit, len(n.log)-1)
		n.applyLogs()
	}

	reply.Success = true
}

func (n *Node) SubmitCommand(cmd interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != Leader || n.crashed {
		return
	}

	n.log = append(n.log, LogEntry{
		Term:    n.currentTerm,
		Command: cmd,
	})
	n.broadcastAppendEntries()
}

func (n *Node) lastLogInfo() (int, int) {
	if len(n.log) == 0 {
		return -1, -1
	}
	return len(n.log) - 1, n.log[len(n.log)-1].Term
}

func (n *Node) isLogUpToDate(lastLogIndex, lastLogTerm int) bool {
	if len(n.log) == 0 {
		return true
	}
	lastTerm := n.log[len(n.log)-1].Term
	return lastLogTerm > lastTerm || (lastLogTerm == lastTerm && lastLogIndex >= len(n.log)-1)
}

func (n *Node) stepDown(term int) {
	n.state = Follower
	n.currentTerm = term
	n.votedFor = -1
	n.resetElectionTimer()
}

func (n *Node) becomeLeader() {
	if n.crashed {
		return
	}
	n.state = Leader
	n.leaderID = n.ID
	n.nextIndex = make(map[int]int)
	n.matchIndex = make(map[int]int)

	for _, peer := range n.peers {
		n.nextIndex[peer.ID] = len(n.log)
		n.matchIndex[peer.ID] = 0
	}

	n.startHeartbeats()
}

func (n *Node) startHeartbeats() {
	if n.heartbeatTimer != nil {
		n.heartbeatTimer.Stop()
	}
	n.heartbeatTimer = time.AfterFunc(50*time.Millisecond, func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		if n.state != Leader || n.crashed {
			return
		}
		n.broadcastAppendEntries()
		n.startHeartbeats()
	})
}

func (n *Node) broadcastAppendEntries() {
	for _, peer := range n.peers {
		go func(p *Node) {
			n.mu.Lock()
			if n.crashed || n.networkPartitioned[p.ID] {
				n.mu.Unlock()
				return
			}
			nextIndex := n.nextIndex[p.ID]
			prevLogIndex := nextIndex - 1
			prevLogTerm := -1
			if prevLogIndex >= 0 && prevLogIndex < len(n.log) {
				prevLogTerm = n.log[prevLogIndex].Term
			}
			entries := n.log[nextIndex:]
			args := AppendEntriesArgs{
				Term:         n.currentTerm,
				LeaderID:     n.ID,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      entries,
				LeaderCommit: n.commitIndex,
			}

			n.mu.Unlock()

			if p.crashed {
				return
			}
			var reply AppendEntriesReply
			p.HandleAppendEntries(&args, &reply)

			n.mu.Lock()
			defer n.mu.Unlock()

			if reply.Term > n.currentTerm {
				n.stepDown(reply.Term)
				return
			}
			if reply.Success {
				n.nextIndex[p.ID] = nextIndex + len(entries)
				n.matchIndex[p.ID] = n.nextIndex[p.ID] - 1
				n.updateCommitIndex()
			} else {
				if n.nextIndex[p.ID] > 0 {
					n.nextIndex[p.ID]--
				}
			}
		}(peer)
	}
}

func (n *Node) updateCommitIndex() {
	for N := len(n.log) - 1; N > n.commitIndex; N-- {
		if n.log[N].Term != n.currentTerm {
			continue
		}
		count := 1
		for _, peer := range n.peers {
			if n.matchIndex[peer.ID] >= N {
				count++
			}
		}
		if count >= majority(len(n.peers)+1) {
			n.commitIndex = N
			n.applyLogs()
			break
		}
	}
}

func (n *Node) applyLogs() {
	for n.lastApplied < n.commitIndex {
		n.lastApplied++
		entry := n.log[n.lastApplied]
		fmt.Printf("Node %d applied command: %v\n", n.ID, entry.Command)
		n.applyCh <- ApplyMsg{
			CommandValid: true,
			Command:      entry.Command,
			CommandIndex: n.lastApplied,
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func majority(n int) int {
	return (n / 2) + 1
}

func main() {
	rand.Seed(time.Now().UnixNano())

	nodes := make([]*Node, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = NewNode(i)
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i != j {
				nodes[i].peers = append(nodes[i].peers, nodes[j])
			}
		}
	}

	for _, node := range nodes {
		go node.Start()
	}

	app := tview.NewApplication()
	table := tview.NewTable().SetFixed(1, 0).SetSelectable(false, false)

	headers := []string{"Node", "State", "Term", "CommitIdx", "LogLen", "LeaderID", "Status", "Partitions"}
	for i, h := range headers {
		table.SetCell(0, i,
			tview.NewTableCell(h).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}

	stateNames := []string{"Follower", "Candidate", "Leader"}

	var commandInput *tview.InputField

	commandInput = tview.NewInputField().
		SetLabel("Command to Leader: ").
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				command := commandInput.GetText()
				if command != "" {
					for _, node := range nodes {
						node.mu.Lock()
						if node.state == Leader && !node.crashed {
							node.SubmitCommand(command)
							node.mu.Unlock()
							break
						}
						node.mu.Unlock()
					}
				}
				commandInput.SetText("")
			}
		})

	inputBuffer := ""
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		ch := event.Rune()

		if ch == 'p' {
			inputBuffer = "p"
			return nil
		}
		if len(inputBuffer) == 1 && inputBuffer == "p" {
			inputBuffer += string(ch)
			return nil
		}
		if len(inputBuffer) == 2 {
			inputBuffer += string(ch)
			a := int(inputBuffer[1] - '0')
			b := int(inputBuffer[2] - '0')
			if a >= 0 && a < 5 && b >= 0 && b < 5 && a != b {
				nodes[a].mu.Lock()
				nodes[b].mu.Lock()
				toggled := !nodes[a].networkPartitioned[b]
				nodes[a].networkPartitioned[b] = toggled
				nodes[b].networkPartitioned[a] = toggled
				nodes[a].mu.Unlock()
				nodes[b].mu.Unlock()
			}
			inputBuffer = ""
			return nil
		}

		if ch >= '0' && ch <= '4' {
			id := int(ch - '0')
			node := nodes[id]
			node.mu.Lock()
			node.crashed = !node.crashed
			if !node.crashed {
				node.resetElectionTimer()
				if node.state == Leader {
					node.startHeartbeats()
				}
			}
			node.mu.Unlock()
		}
		return event
	})

	go func() {
		for {
			for i, node := range nodes {
				node.mu.Lock()
				state := stateNames[node.state]
				term := node.currentTerm
				commitIdx := node.commitIndex
				logLen := len(node.log)
				leaderID := node.leaderID
				status := "Up"
				if node.crashed {
					status = "Down"
				}
				partitionStr := ""
				for pid := range node.networkPartitioned {
					partitionStr += fmt.Sprintf("%d ", pid)
				}
				node.mu.Unlock()

				table.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("%d", node.ID)).SetAlign(tview.AlignCenter))
				table.SetCell(i+1, 1, tview.NewTableCell(state).SetAlign(tview.AlignCenter))
				table.SetCell(i+1, 2, tview.NewTableCell(fmt.Sprintf("%d", term)).SetAlign(tview.AlignCenter))
				table.SetCell(i+1, 3, tview.NewTableCell(fmt.Sprintf("%d", commitIdx)).SetAlign(tview.AlignCenter))
				table.SetCell(i+1, 4, tview.NewTableCell(fmt.Sprintf("%d", logLen)).SetAlign(tview.AlignCenter))
				table.SetCell(i+1, 5, tview.NewTableCell(fmt.Sprintf("%d", leaderID)).SetAlign(tview.AlignCenter))
				table.SetCell(i+1, 6, tview.NewTableCell(status).SetAlign(tview.AlignCenter))
				table.SetCell(i+1, 7, tview.NewTableCell(partitionStr).SetAlign(tview.AlignCenter))
			}
			app.Draw()
			time.Sleep(300 * time.Millisecond)
		}
	}()

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, false).
		AddItem(commandInput, 1, 1, true).
		AddItem(tview.NewTextView().
			SetText("Press 0-4 to toggle node crash/recover | Press 'pXY' to partition X and Y | Enter command and press Enter | Ctrl+C to quit").
			SetTextColor(tcell.ColorGreen).
			SetTextAlign(tview.AlignCenter), 1, 1, false)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
