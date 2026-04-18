package main

import (
	"fmt"
	"log"
	"neuronfs/pose_converter"
)

func HelloWorld() string {
	return "Hello, World!"
}

func main() {
	fmt.Println(HelloWorld())
	
	// Example Pose Conversion via pose_converter package
	poseJson, err := pose_converter.ConvertPoseData("shoulder_right", 10.5, 20.1, 5.0)
	if err != nil {
		log.Fatalf("Error converting pose: %v", err)
	}
	fmt.Println("Converted Pose:", poseJson)
}
