#!/bin/bash

if ! command -v tmux &> /dev/null; then
    echo "Error: tmux is not installed. Please install it first (e.g. 'sudo apt install tmux' or 'brew install tmux') to start (Seat)surfing 🏄."
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$*" == *"--install"* ]]; then
    (cd "$SCRIPT_DIR/ui" && npm ci)
fi

RUN_SERVER_CMD="./run.sh"
if [[ "$*" == *"--clear-db"* ]]; then
    RUN_SERVER_CMD="./run.sh --clear-db"
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
tmux send-keys -t "$SESSION:0.2" "cd '$SCRIPT_DIR/server' && $RUN_SERVER_CMD" Enter

# Bottom: dev console (full width)
tmux split-window -v -f -t "$SESSION:0"

CONSOLE_CMD="\
cd '$SCRIPT_DIR' \
&& add-missing-translations() { \
  tmux send-keys -t '$SESSION:0.0' C-c '' Enter \
  && sleep 1 \
  && tmux send-keys -t '$SESSION:0.0' \"cd '$SCRIPT_DIR/ui' && ./add-missing-translations.sh && npm run dev\" Enter; \
} \
&& export -f add-missing-translations \
&& prettierFormat() { \
  tmux send-keys -t '$SESSION:0.0' C-c '' Enter \
  && sleep 1 \
  && tmux send-keys -t '$SESSION:0.0' \"cd '$SCRIPT_DIR/ui' && npm run prettier:format && npm run dev\" Enter; \
} \
&& export -f prettierFormat \
&& restartServer() { \
  tmux send-keys -t '$SESSION:0.2' C-c '' Enter \
  && sleep 1 \
  && tmux send-keys -t '$SESSION:0.2' \"cd '$SCRIPT_DIR/server' && ./run.sh\" Enter; \
} \
&& export -f restartServer \
&& printf '\nLogin: http://localhost:3000/ui/ (user: admin@seatsurfing.local / password: Sea!surf1ng)\nMails: http://localhost:8025\n\nCommands:\n- restartServer: restarts the backend server\n- prettierFormat: runs code formatting\n- add-missing-translations: adds missing translations\n\nTip: run ./dev.sh --clear-db to start with a clean, empty database.\n\nHappy Seatsurfing … 🏄\n\n' \
; bash \
; tmux kill-session -t '$SESSION'\
"

tmux send-keys -t "$SESSION:0.3" "$CONSOLE_CMD" Enter

tmux attach-session -t "$SESSION"
