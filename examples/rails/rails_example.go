// This exmaple shows how you might put a falcore server in front
// of a rails app running in something like thin or unicorn (instead of, say, nginx).
// It will try to serve files from -public and then fall through
// to proxying the request to a list of ports on localhost
//
// To test this simply:
// 		cd $GOPATH/src/github.com/fitstar/falcore
// 		mkdir bin
// 		go build -o bin/hello ./examples/hello_world
// 		go build -o bin/rails ./examples/rails
// 		bin/hello -port 3000 &
// 		bin/rails -public ./test
// Here we're pretending hello is a rails server.  If you happen to have a rails
// app handy, you can use that instead.  Now direct your browser to http://localhost:3001.
// You will see the index.html page from ./test.  Any file path that doesn't have
// a matching file in ./test will return the hello world response from the upstream
// server.

package main

import (
	"flag"
	"fmt"
	"github.com/fitstar/falcore"
	"github.com/fitstar/falcore/filter"
	"regexp"
	"strconv"
)

var (
	flagPort     = flag.Int("port", 3001, "the port to listen on")
	flagPath     = flag.String("public", "", "the path to the 'public' directory of the app")
	flagUpstream = flag.String("up", "3000", "comma separated list of upstream ports")
)

func main() {
	// parse command line options
	flag.Parse()

	// create pipeline
	pipeline := falcore.NewPipeline()

	// setup file server for public directory
	if *flagPath != "" {
		// Serve files from public directory
		pipeline.Upstream.PushBack(&filter.FileFilter{
			BasePath:       *flagPath,
			DirectoryIndex: "index.html",
		})
	} else {
		falcore.Warn("Path to public directory is missing")
	}

	// parse upstream list and create the upstream pool
	upStrings := regexp.MustCompile("[0-9]+").FindAllString(*flagUpstream, -1)
	pool := filter.NewUpstreamPool("railsdemo")
	for _, s := range upStrings {
		port, _ := strconv.Atoi(s)
		pool.AddUpstream(filter.NewUpstream(filter.NewUpstreamTransport("localhost", port, 0, nil)), 1)
	}

	// create upstream pool and add to pipeline
	if len(upStrings) > 0 {
		pipeline.Upstream.PushBack(pool)
	} else {
		falcore.Warn("No upstream ports provided")
	}

	// add any downstream filters you might want such as etag support or compression

	// setup server
	server := falcore.NewServer(*flagPort, pipeline)

	// start the server
	// this is normally blocking forever unless you send lifecycle commands
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Could not start server:", err)
	}

}
