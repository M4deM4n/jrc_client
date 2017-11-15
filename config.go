package main

type Config struct {
	handle     string
	host       string
	port       int
	timeStamps bool
	allowPush  bool
	pushURLs   bool
	msgSound   bool
}

func DefaultConfig() Config {
	conf := Config{}
	conf.handle = "nobody"
	conf.host = "127.0.0.1"
	conf.port = 6000
	conf.allowPush = true
	conf.pushURLs = true
	conf.msgSound = true
	conf.timeStamps = true
	return conf
}
