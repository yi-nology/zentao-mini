# Makefile for chandao-mini backend

.PHONY: all build run clean

# 构建目标
all: build

build: 
	go build -o server .

run: build
	./server

clean:
	rm -f server
