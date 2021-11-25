# DISYSMP3

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative Auction/Auction.proto

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative Replica/Replica.proto

# How to run 
* start 2 or more Servers upto 10 from /Server go run .
* start 1 front end from /FrontEnd go run .
    * The auction ends 3 minutes after frontend starts  
* start 1 or more clients from /Client go run .
    * client CLI commands
    * "bid" which promts for an integer
    * "result" prints the highest bid and bidder

# Be aware 
* restarting the front end restart the auction timer but doesnt clear the logs
* previous logs will be included, so delete these before starting a new auction   
