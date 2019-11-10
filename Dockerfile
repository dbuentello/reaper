FROM golang:stretch
MAINTAINER @audibleblink

# Build the Docker image first
#  > sudo docker build -t reaper .

# To just generate reaper binaries, run the following and check your `src` folder for the output
#  > sudo docker run --rm --mount type=bind,src=/tmp,dst=/go/src/github.com/infosechoudini/reaper//data/temp reaper make linux
#  > ls /tmp/v0.6.4.BETA

# To start the reaper Server, run
#  > sudo docker run -it -p 443:443 reaper


RUN apt-get update && apt-get install -y git make
RUN go get github.com/infosechoudini/reaper/...

WORKDIR $GOPATH/srcgithub.com/infosechoudini/reaper/
VOLUME ["data/temp"]
EXPOSE 443
CMD ["go", "run", "cmd/reaperserver/main.go", "-i", "0.0.0.0"]
