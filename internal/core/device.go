package core

import "runtime"

func GetPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		panic("Unsupported platform: " + runtime.GOOS)
	}
}
