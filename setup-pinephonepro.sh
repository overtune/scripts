#!/bin/bash

# TODO: Only run commands once

# Console colors
OK='\033[0;32m'
WARNING="\033[93m"
ERROR="\033[91m"
CHECKMARK="\xE2\x9C\x94"

# Variables
use_pacman='true'
use_snap='true'
use_ohmyzsh='true'
use_dotfiles='true'
use_node='true'
skip=''

print_usage() {
  printf "Usage: ..."
}

# Handle input props
while getopts s: flag;
do
  case "${flag}" in
	s) IFS=', ' read -r -a skip <<< "${OPTARG}"
  esac
done

# Go home
cd $HOME

# Install software
if [[ ! " ${skip[*]} " =~ " pacman " ]]; then
	printf "${WARNING}1/5 Installning pacman packages\n"
	sudo pacman -S --needed vim tmux git zsh chromium snapd
	printf "${OK}${CHECKMARK} 1/5 Installning pacman packages\n"
else
	printf "${WARNING}${CHECKMARK} 1/5 Skipped: Installning pacman packages\n"
fi

# Enable snap
if [[ ! " ${skip[*]} " =~ " snap " ]]; then
	printf "${WARNING}2/5 Enabling Snap\n"
	sudo systemctl enable --now snapd.socket
	printf "${OK}${CHECKMARK} 2/5 Enabling Snap\n"
else
	printf "${WARNING}${CHECKMARK} 2/5 Skipped: Enabling Snap\n"
fi

# Install ohmyzsh
if [[ ! " ${skip[*]} " =~ " ohmyzsh " ]]; then
	printf "${WARNING}3/5 Setting up OhMyZsh\n"
	sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" 
	printf "${OK}${CHECKMARK} 3/5 Setting up OhMyZsh\n"
else
	printf "${WARNING}${CHECKMARK} 3/5 Skipped: Setting up OhMyZsh\n"
fi

# Setup dotfiles
if [[ ! " ${skip[*]} " =~ " dotfiles " ]]; then
	printf "${WARNING}4/5 Setting up dotfiles\n"
	git clone --bare https://github.com/overtune/dotfiles.git $HOME/.dotfiles
	alias config='/usr/bin/git --git-dir=$HOME/.dotfiles/ --work-tree=$HOME'
	config checkout
	printf "${OK}${CHECKMARK} 4/5 Setting up dotfiles\n"
else
	printf "${WARNING}${CHECKMARK} 4/5 Skipped: Setting up dotfiles\n"
fi

# Install node.js/nvm
if [[ ! " ${skip[*]} " =~ " node " ]]; then
	printf "${WARNING}5/5 Installing node.js/nvm\n"
	curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash
	source ~/.zshrc
	export NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"
	[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh" # This loads nvm
	nvm install node
	printf "${OK}${CHECKMARK} 5/5 Installing node.js/nvm\n"
else
	printf "${WARNING}${CHECKMARK} 5/5 Skipped: Installing node.js/nvm\n"
fi

# Install snap packages
# sudo snap install supertuxkart
