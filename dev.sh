#!/bin/bash

if ! command -v tmux &> /dev/null; then
    echo "Error: tmux is not installed. Please install it first (e.g. 'sudo apt install tmux' or 'brew install tmux') to start (Seat)surfing 🏄."
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$*" == *"--install"* ]]; then
    (cd "$SCRIPT_DIR/ui" && npm ci)
fi

SESSION="seatsurfing-dev"

tmux new-session -d -s "$SESSION" -x 220 -y 50

tmux split-window -h -t "$SESSION:0.0"
tmux split-window -h -t "$SESSION:0.0"

# Top left: frontend
tmux send-keys -t "$SESSION:0.0" "cd '$SCRIPT_DIR/ui' && npm run dev" Enter

# Top middle: frontend tests
tmux send-keys -t "$SESSION:0.1" "cd '$SCRIPT_DIR/ui' && npm run test" Enter

# Top right: backend
tmux send-keys -t "$SESSION:0.2" "cd '$SCRIPT_DIR/server' && ./run.sh" Enter

# Bottom: dev console (full width)
tmux split-window -v -f -t "$SESSION:0"

CONSOLE_CMD="\
cd '$SCRIPT_DIR' \
&& restartServer() { \
  tmux send-keys -t '$SESSION:0.2' C-c '' Enter \
  && sleep 1 \
  && tmux send-keys -t '$SESSION:0.2' \"cd '$SCRIPT_DIR/server' && ./run.sh\" Enter; \
} \
&& export -f restartServer \
&& clearDatabase() { \
  tmux send-keys -t '$SESSION:0.2' C-c '' Enter \
  && sleep 1 \
  && docker exec postgres-seatsurfing psql -U postgres -c 'DROP DATABASE IF EXISTS seatsurfing;' \
  && docker exec postgres-seatsurfing psql -U postgres -c 'CREATE DATABASE seatsurfing;' \
  && tmux send-keys -t '$SESSION:0.2' \"cd '$SCRIPT_DIR/server' && ./run.sh\" Enter; \
} \
&& export -f clearDatabase \
&& printf '\nLogin: http://localhost:3000/ui/ (user: admin@seatsurfing.local / password: Sea!surf1ng)\nMails: http://localhost:8025\n\nCommands:\n- restartServer: restarts the backend server\n- clearDatabase: restarts with a clean new database\n\nHappy Seatsurfing … 🏄\n\n' \
; bash \
; tmux kill-session -t '$SESSION'\
"

tmux send-keys -t "$SESSION:0.3" "$CONSOLE_CMD" Enter

tmux attach-session -t "$SESSION"
