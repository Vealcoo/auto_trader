FROM golang:latest
RUN mkdir /root/.config
ADD .config /root/.config
ENV SRC_DIR=${PWD}:cmd
COPY . ${SRC_DIR}
WORKDIR ${SRC_DIR}
EXPOSE 8888
CMD go run cmd/main.go -p 8888
