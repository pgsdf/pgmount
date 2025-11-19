package notify

import (
	"fmt"
	"os/exec"
	"strconv"
)

var initialized bool

// Init initializes the notification system
func Init() error {
	// Check if notify-send is available
	_, err := exec.LookPath("notify-send")
	if err != nil {
		return fmt.Errorf("notify-send not found in PATH (install libnotify)")
	}
	
	initialized = true
	return nil
}

// Close closes the notification system
func Close() {
	initialized = false
}

// Send sends a desktop notification
func Send(summary, body string, timeout int) error {
	return SendWithIcon(summary, body, "drive-removable-media", timeout)
}

// SendWithIcon sends a desktop notification with a custom icon
func SendWithIcon(summary, body, icon string, timeout int) error {
	if !initialized {
		return fmt.Errorf("notification system not initialized")
	}

	args := []string{}
	
	// Add timeout if specified
	if timeout > 0 {
		args = append(args, "-t", strconv.Itoa(timeout))
	}
	
	// Add icon
	if icon != "" {
		args = append(args, "-i", icon)
	}
	
	// Add summary and body
	args = append(args, summary, body)
	
	cmd := exec.Command("notify-send", args...)
	return cmd.Run()
}
