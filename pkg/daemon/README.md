# daemon

Package daemon contains code allowing daemon process creation.

## Usage
The deamon process is spawned using the package [go-deamon](https://github.com/sevlyar/go-daemon/tree/master).
To spawn a daemon process, first define its entry point and CLI arguments by implementing the `Daemon` interface. 
Make sure to implement the `Serialize` method, so that arguments can be properly passed to the daemon.