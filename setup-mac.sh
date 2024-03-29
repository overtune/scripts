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
# Check for Homebrew, install if we don't have it
if test ! $(which brew); then
	printf "${OK}Install Homebrew\n"
    ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
fi

# Install brews
# printf "${OK}Install git\n"
# brew install git
printf "${OK}Install neovim\n"
brew install vim
printf "${OK}Install tmux\n"
brew install tmux
printf "${OK}Install node\n"
brew install node
printf "${OK}Install python3\n"
brew install python@3.9
printf "${OK}Install thefuck\n"
brew install thefuck
printf "${OK}Install ack\n"
brew install ack
printf "${OK}Install fff\n"
brew install fff
printf "${OK}Install fzf\n"
brew install fzf
printf "${OK}Install lazygit\n"
brew install jesseduffield/lazygit/lazygit
printf "${OK}Install lazydocker\n"
brew install jesseduffield/lazydocker/lazydocker
printf "${OK}Install reattach-to-user-namespace\n"
brew install reattach-to-user-namespace
$(brew --prefix)/opt/fzf/install
printf "${OK}Install iterm2\n"
brew install --cask iterm2
# printf "${OK}Install alacritty\n"
# brew install --cask alacritty
printf "${OK}Install docker\n"
brew install --cask docker
printf "${OK}Install qmktoolbox\n"
brew tap homebrew/cask-drivers
brew install --cask qmk-toolbox

# Setup neovim 
printf "${OK}Install neovim dependencies\n"
npm install -g neovim
python3 -m pip install --user --upgrade pynvim

# Install mas 
# printf "${OK}Install mas\n"
# brew install mas
# mas signin $email
# printf "${OK}Install iA Writer\n"
# mas install 775737590 # iA Writer
# printf "${OK}Install Lastpass\n"
# mas install 926036361 # Lastpass
# printf "${OK}Install Magnet\n"
# mas install 441258766 # Magnet
