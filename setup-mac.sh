#!/bin/bash

# Console colors
OK='\033[0;32m'
WARNING="\033[93m"
ERROR="\033[91m"

# Get user email 
printf "${WARNING}Please enter your AppleID Email address\n"
read email

# Setup folder structure
printf "${OK}Setup folder structure\n"
mkdir -p  ~/Development/work 
mkdir -p  ~/Development/private

# Install Xcode and command line tools
printf "${OK}Install Xcode and command line tools\n"
xcode-select --install

# Install Homebrew 
printf "${OK}Install Homebrew\n"
curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh 

# Install brews
# printf "${OK}Install git\n"
# brew install git
printf "${OK}Install vim\n"
brew install vim
printf "${OK}Install tmux\n"
brew install tmux
printf "${OK}Install nodejs\n"
brew install nodejs
printf "${OK}Install python3\n"
brew install python3
printf "${OK}Install alacritty\n"
brew install alacritty
printf "${OK}Install docker\n"
brew install docker
printf "${OK}Install qmktoolbox\n"
brew install qmktoolbox
printf "${OK}Install thefuck\n"
brew install thefuck
printf "${OK}Install ack\n"
brew install ack
printf "${OK}Install fff\n"
brew install fff
printf "${OK}Install fzf\n"
brew install fzf
$(brew --prefix)/opt/fzf/install

# Install mas 
printf "${OK}Install mas\n"
brew install mas
mas signin $email
printf "${OK}Install iA Writer\n"
mas install 775737590 # iA Writer
printf "${OK}Install Lastpass\n"
mas install 926036361 # Lastpass
printf "${OK}Install Magnet\n"
mas install 441258766 # Magnet
