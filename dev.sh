#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

SESSION="seatsurfing-dev"

tmux new-session -d -s "$SESSION" -x 220 -y 50

# Pane 0 (links): UI
tmux send-keys -t "$SESSION:0.0" "cd '$SCRIPT_DIR/ui' && npm ci && npm run dev" Enter

# Pane 1 (rechts oben): Server
tmux split-window -h -t "$SESSION:0.0"
tmux send-keys -t "$SESSION:0.1" "cd '$SCRIPT_DIR/server' && ./run.sh" Enter

# Pane 2 (rechts unten): Arbeits-Terminal — beim Verlassen wird die Session gekillt
tmux split-window -v -t "$SESSION:0.1"
tmux send-keys -t "$SESSION:0.2" "cd '$SCRIPT_DIR'; bash; tmux kill-session -t '$SESSION'" Enter

tmux attach-session -t "$SESSION"
