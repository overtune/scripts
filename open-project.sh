#!/bin/bash

PROJECT_DIR=$1
DIR=~/Development/work/$PROJECT_DIR

if [ -d "$DIR" ]; then
	cd $DIR;
	set -- $(stty size) # $1 = rows $2 = columns
	tmux -2 new-session -d -s "$PROJECT_DIR" -x "$2" -y "$(($1 - 1))" # status line uses a row
	tmux new -d -s $PROJECT_DIR
	tmux split-window -h -p 27 -t $PROJECT_DIR
	tmux split-window -v  -t $PROJECT_DIR
	tmux select-pane -L -t $PROJECT_DIR
	tmux send-keys -t $PROJECT_DIR "vim" Enter
	tmux attach -t $PROJECT_DIR
else
	echo "Error: ${DIR} not found. Can not continue."
	exit 1
fi
