package main

import (
	"fmt"
	"neuronfs/pose_converter"
)

func HelloWorld() string {
	return "Hello, World!"
}

func main() {
	fmt.Println(HelloWorld())
	
	// Example Pose Conversion via pose_converter package
	poseJson, _ := pose_converter.ConvertPoseData("shoulder_right", 10.5, 20.1, 5.0)
	fmt.Println("Converted Pose:", poseJson)
}
