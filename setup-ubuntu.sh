#!/bin/bash

# Define props
mode=${mode:-base}

# Handle input props
while [ $# -gt 0 ]; do
   if [[ $1 == *"--"* ]]; then
        param="${1/--/}"
        declare $param="$2"
   fi
  shift
done

# Update and install software
sudo apt update
sudo add-apt-repository universe
sudo apt install vim tmux git zsh mosh build-essential cmake python3-dev -y

# Install Oh-my-shell
curl -O https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh
sh install.sh
rm install.sh

# Get the dotfiles
cd ~
git clone https://github.com/overtune/dotfiles.git
cd dotfiles
bash makesymlinks.sh

# Install conditional software
if [ $mode = "full" ]
then
	# Golang
	curl -O https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
	tar -C /usr/local -xzf go1.11.2.linux-amd64.tar.gz
	export PATH=$PATH:/usr/local/go/bin

	# Node
	export NVM_DIR="$HOME/.nvm" && (
	  git clone https://github.com/creationix/nvm.git "$NVM_DIR"
	  cd "$NVM_DIR"
	  git checkout `git describe --abbrev=0 --tags --match "v[0-9]*" $(git rev-list --tags --max-count=1)`
	) && \. "$NVM_DIR/nvm.sh"
	nvm install node
fi

# Install Vim plugs
vim +PlugInstall +qall +silent
cd ~/.vim/plugged/YouCompleteMe && python3 install.py 
