
#!/bin/bash

echo "Downloading bench..."
curl -L "https://github.com/Shravan-1908/bench/releases/latest/download/bench-darwin-amd64" -o bench

echo "Adding bench into PATH..."

mkdir -p ~/.bench;
mv ./bench ~/.bench
echo "export PATH=$PATH:~/.bench" >> ~/.bashrc

echo "bench installation is completed!"
echo "You need to restart the shell to use bench."
