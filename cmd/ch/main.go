package main

import "github.com/timo-reymann/ContainerHive/internal/buildinfo"

func main() {
	println("ContainerHive CLI")
	buildinfo.PrintVersionInfo()
}
