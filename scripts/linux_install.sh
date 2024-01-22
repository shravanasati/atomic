#!/bin/bash

echo "Downloading atomic..."
curl -L "https://github.com/shravanasati/atomic/releases/latest/download/atomic-linux-amd64" -o atomic

echo "Adding atomic into PATH..."

mkdir -p ~/.atomic

chmod u+x ./atomic

mv ./atomic ~/.atomic
echo "export PATH=$PATH:~/.atomic" >> ~/.bashrc

echo "atomic installation is completed!"
echo "You need to restart the shell to use atomic."
