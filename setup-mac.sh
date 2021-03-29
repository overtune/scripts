#!/bin/bash

# Define props
# email=${email}
# skip=${skip}

# Handle input props
while [ $# -gt 0 ]; do
   if [[ $1 == *"--"* ]]; then
        param="${1/--/}"
        declare $param="$2"
   fi
  shift
done

# Ask for the administrator password upfront
sudo -v

# Install homebrew
if [ "$skip" -ne "homebrew" ]; then
	curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh
else
	echo "Skipping homebrew"
fi

# Install MAS support
brew install mas
mas signin $email 
