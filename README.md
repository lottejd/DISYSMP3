# DISYSMP3
# How to run 
* start 2 or more Servers up to 10 from /Server go run .
* start 1 front end from /FrontEnd go run . 
* start 1 or more clients from /Client go run .
    * client CLI commands
    * "bid" which prompts for an integer
    * "result" prints the highest bid and bidder

# Be aware 
* The auction ends 3 minutes after the front end starts 
* restarting the front end restarts the auction timer, but doesn't clear the logs
* previous logs will be included when front end starts, so delete these before starting a new auction   

# Proto commands

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative Auction/Auction.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative Replica/Replica.proto
